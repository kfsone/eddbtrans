package eddbtrans

import (
	"errors"
	"sync"
)

// Declare an error specifically for denoting the registry is closed.
var ErrDaycareClosed = errors.New("registry is already closed")

// Utility for allowing out-of-order registration and lookup of parent<->child relations,
// for example querying whether the system id of a station is valid, without knowing
// whether the system will have been loaded yet.

type parentCheck struct {
	parentID uint64
	entity   interface{}
}

// Daycare is a place for parents to register and children to look them up.
type Daycare struct {
	registrations chan uint64
	inquiries     chan parentCheck
	approvals     chan interface{}
	working       sync.WaitGroup

	// statistics
	Registered uint64
	Queried    uint64
	Approved   uint64
	Queued     uint64
	Duplicate  uint64
}

// Close releases memory used by the Daycare once the registrations and inquiries
// channels have been closed.
func (dc *Daycare) Close() (err error) {
	if dc.Closed() {
		err = ErrDaycareClosed
	} else {
		// Wait for the worker to shutdown.
		dc.working.Wait()
		// Clear out the values so we can detect repeat calls.
		dc.registrations, dc.inquiries, dc.approvals = nil, nil, nil
	}
	return
}

// Closed returns true if the Daycare is not currently open.
func (dc Daycare) Closed() bool {
	return dc.registrations == nil
}

// Approvals returns a channel from which approvals may be consumed.
func (dc Daycare) Approvals() <-chan interface{} {
	return dc.approvals
}

// CloseInquiries closes down the inquiry channel.
func (dc *Daycare) CloseInquiries() {
	if !dc.Closed() {
		close(dc.inquiries)
	}
}

// CloseRegistrations closes down the registration channel.
func (dc *Daycare) CloseRegistrations() {
	if !dc.Closed() {
		close(dc.registrations)
	}
}

// Query forwards an entity to be matched against its parent by the Daycare center.
func (dc *Daycare) Query(parentID uint64, entity interface{}) (err error) {
	if dc.Closed() {
		err = ErrDaycareClosed
	} else {
		dc.inquiries <- parentCheck{parentID: parentID, entity: entity}
		dc.Queried++
	}
	return
}

// Register adds an entity to the registry, approving any pending and future queries.
func (dc *Daycare) Register(entityID uint64) (err error) {
	if dc.Closed() {
		err = ErrDaycareClosed
	} else {
		dc.registrations <- entityID
		dc.Registered++
	}
	return
}

// OpenDayCare launches a worker which will monitor incoming registrations and inquiries,
// which will be processed through a registry. Inquiries for parents not yet registered are
// placed into a waiting queue. When registrations come in, they check if they have a queue
// and if so, immediately approve the waiters in the order they arrived.
func OpenDayCare() (dc *Daycare) {
	// Create the instance with channels.
	dc = &Daycare{
		registrations: make(chan uint64, 8),
		inquiries:     make(chan parentCheck, 1),
		approvals:     make(chan interface{}, 64),
	}

	// Start the background worker.
	dc.working.Add(1)
	go func() {
		defer dc.working.Done()
		defer close(dc.approvals)

		registry := make(map[uint64][]interface{})

		channelsOpen := 2

		for channelsOpen > 0 {
			select {

			// Receive an incoming registration.
			case id, ok := <-dc.registrations:
				if ok == false {
					// Channel closed
					channelsOpen--
					continue
				}
				// Register an id if it's not already registered.
				waiting, existed := registry[id]
				if existed {
					// Had children tried to look up this parent?
					if waiting != nil {
						for _, child := range waiting {
							dc.approvals <- child
						}
						dc.Approved += uint64(len(waiting))
					} else {
						dc.Duplicate++
					}
					// Clear the list
				}
				// This will prevent future children from registering
				registry[id] = nil

				break

				// Receive an inquiry (has this ID been registered).
			case check, ok := <-dc.inquiries:
				if ok == false {
					channelsOpen--
					continue
				}
				// See if the parent exists, or register us as waiting for it
				// For the parent to be registered, it should exist but point to nil.
				if waiting, existed := registry[check.parentID]; existed == false {
					registry[check.parentID] = make([]interface{}, 0, 16)
					registry[check.parentID] = append(registry[check.parentID], check.entity)
					dc.Queued++
				} else if waiting != nil {
					// there's a queue
					registry[check.parentID] = append(registry[check.parentID], check.entity)
					dc.Queued++
				} else {
					// registered, approve
					dc.approvals <- check.entity
					dc.Approved++
				}

				break
			}
		}
	}()

	return
}
