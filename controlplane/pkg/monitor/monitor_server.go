package monitor

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	defaultSize            = 10
	UPDATE                 = "UPDATE"
	DELETE                 = "DELETE"
	INITIAL_STATE_TRANSFER = "INITIAL_STATE_TRANSFER"
)

type Entity interface {
	GetId() string
}

type Event struct {
	EventType string
	Entities  map[string]Entity
}

type EventConverter interface {
	Convert(event Event) (interface{}, error)
}

type Recipient interface {
	SendMsg(msg interface{}) error
}

type MonitorServer interface {
	Update(entity Entity)
	Delete(entity Entity)

	AddRecipient(recipient Recipient)
	DeleteRecipient(recipient Recipient)
	MonitorEntities(stream grpc.ServerStream) error

	Serve()
}

type monitorServerImpl struct {
	eventConverter           EventConverter
	eventCh                  chan Event
	newMonitorRecipientCh    chan Recipient
	closedMonitorRecipientCh chan Recipient
	entities                 map[string]Entity
	recipients               []Recipient
}

func NewMonitorServer(eventConverter EventConverter) MonitorServer {
	return &monitorServerImpl{
		eventConverter:           eventConverter,
		eventCh:                  make(chan Event, defaultSize*100),
		newMonitorRecipientCh:    make(chan Recipient, defaultSize),
		closedMonitorRecipientCh: make(chan Recipient, defaultSize),
		entities:                 make(map[string]Entity),
		recipients:               make([]Recipient, 0, defaultSize),
	}
}

func (m *monitorServerImpl) Update(entity Entity) {
	logrus.Infof("MONITOR_UPDATE: %v", len(m.eventCh))
	m.eventCh <- Event{
		EventType: UPDATE,
		Entities:  map[string]Entity{entity.GetId(): entity},
	}
}

func (m *monitorServerImpl) Delete(entity Entity) {
	m.eventCh <- Event{
		EventType: DELETE,
		Entities:  map[string]Entity{entity.GetId(): entity},
	}
}

func (m *monitorServerImpl) AddRecipient(recipient Recipient) {
	logrus.Infof("MonitorServerImpl.AddRecipient: %v", recipient)
	m.newMonitorRecipientCh <- recipient
}

func (m *monitorServerImpl) DeleteRecipient(recipient Recipient) {
	logrus.Infof("MonitorServerImpl.DeleteRecipient: %v", recipient)
	m.closedMonitorRecipientCh <- recipient
}

func (m *monitorServerImpl) MonitorEntities(stream grpc.ServerStream) error {
	m.AddRecipient(stream)

	// We need to wait until it will be done and do not exit
	for {
		select {
		case <-stream.Context().Done():
			m.DeleteRecipient(stream)
			return nil
		}
	}
}

func (m *monitorServerImpl) Serve() {
	logrus.Infof("Serve starting...")
	for {
		select {
		case newRecipient := <-m.newMonitorRecipientCh:
			initialStateTransferEvent := Event{
				EventType: INITIAL_STATE_TRANSFER,
				Entities:  m.entities,
			}
			m.send(initialStateTransferEvent, newRecipient)
			m.recipients = append(m.recipients, newRecipient)
		case closedRecipient := <-m.closedMonitorRecipientCh:
			for j, r := range m.recipients {
				if r == closedRecipient {
					m.recipients = append(m.recipients[:j], m.recipients[j+1:]...)
					break
				}
			}
		case event := <-m.eventCh:
			logrus.Infof("New event: %v", event)
			for _, entity := range event.Entities {
				if event.EventType == UPDATE {
					m.entities[entity.GetId()] = entity
				}
				if event.EventType == DELETE {
					delete(m.entities, entity.GetId())
				}
			}
			m.send(event, m.recipients...)
		}
	}
}

func (m *monitorServerImpl) send(event Event, recipients ...Recipient) {
	for _, recipient := range recipients {
		msg, err := m.eventConverter.Convert(event)
		if err != nil {
			logrus.Errorf("Error during converting event: %v", err)
		}
		if err := recipient.SendMsg(msg); err != nil {
			logrus.Errorf("Error during send: %+v", err)
		}
	}
}
