package sdl

import (
	"context"
	"errors"
	"fmt"
	"unsafe"

	// Packages
	sdl "github.com/veandco/go-sdl2/sdl"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Context struct {
	evt   map[uint32]func(unsafe.Pointer)
	audio map[uint32]*Audio
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new SDL object which can output audio and video
func New(t Type) (*Context, error) {
	var flags uint32
	if t.Is(VIDEO) {
		flags |= sdl.INIT_VIDEO
	}
	if t.Is(AUDIO) {
		flags |= sdl.INIT_AUDIO
	}
	if err := sdl.Init(flags); err != nil {
		return nil, err
	}
	return &Context{
		evt:   make(map[uint32]func(unsafe.Pointer)),
		audio: make(map[uint32]*Audio),
	}, nil
}

func (ctx *Context) Close() error {
	var result error

	// Close audio devices
	for _, audio := range ctx.audio {
		result = errors.Join(result, audio.Close())
	}

	// Quit SDL
	sdl.Quit()

	// Return any errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Register an event handler and return the event identifier
func (ctx *Context) Register(fn func(userInfo unsafe.Pointer)) uint32 {
	evt := sdl.RegisterEvents(1)
	ctx.evt[evt] = fn
	return evt
}

// Post an event to the event queue with user information
func (ctx *Context) Post(evt uint32, userInfo unsafe.Pointer) {
	sdl.PushEvent(&sdl.UserEvent{
		Type:  evt,
		Data1: userInfo,
	})
}

// Run the event loop
func (ctx *Context) Run(parent context.Context) error {
	// Register an event which quits the application when context is cancelled
	evtCancel := sdl.RegisterEvents(1)
	go func() {
		<-parent.Done()
		sdl.PushEvent(&sdl.UserEvent{
			Type: evtCancel,
		})
	}()

	// Unpause the audio capture/playback devices
	for _, audio := range ctx.audio {
		// Start capturing audio
		sdl.PauseAudioDevice(audio.device, false)
	}

	// Start the runloop
	quit := false
	for {
		if quit {
			break
		}

		// Wait on an event
		evt := sdl.WaitEvent()

		// Handle cancel, custom, keyboard and quit events
		switch evt := evt.(type) {
		case *sdl.UserEvent:
			if evt.GetType() == evtCancel {
				quit = true
			} else if fn, exists := ctx.evt[evt.GetType()]; exists {
				fn(evt.Data1)
			}
		case *sdl.QuitEvent:
			quit = true
		default:
			fmt.Println("Unhandled event:", evt.GetType())
		}
	}

	// Pause the audio capture/playback devices
	for _, audio := range ctx.audio {
		// Start capturing audio
		sdl.PauseAudioDevice(audio.device, true)
	}

	// TODO: Return any errors
	return nil
}
