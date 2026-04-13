package gst

import (
	"errors"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
)

type busWatch struct {
	bus  *gst.Bus
	loop *glib.MainLoop
	done chan struct{}
}

func startBusWatch(bus *gst.Bus, handler gst.BusWatchFunc) (*busWatch, error) {
	if bus == nil {
		return nil, errors.New("gstreamer bus is nil")
	}
	if !bus.AddWatch(handler) {
		return nil, errors.New("failed to add gstreamer bus watch")
	}
	watch := &busWatch{
		bus:  bus,
		loop: glib.NewMainLoop(glib.MainContextDefault(), false),
		done: make(chan struct{}),
	}
	go func() {
		defer close(watch.done)
		watch.loop.Run()
	}()
	return watch, nil
}

func (w *busWatch) Stop() {
	if w == nil {
		return
	}
	if w.bus != nil {
		_ = w.bus.RemoveWatch()
	}
	if w.loop != nil {
		w.loop.Quit()
	}
	if w.done != nil {
		<-w.done
	}
}
