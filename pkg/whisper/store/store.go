package store

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	// Packages
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	whisper "github.com/mutablelogic/go-whisper/sys/whisper"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Store struct {
	sync.RWMutex

	// Path to the models directory and file extension
	path, ext string

	// list of all models
	models []*schema.Model

	// download models
	client whisper.Client
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new model store
func NewStore(path, ext, modelUrl string) (*Store, error) {
	store := new(Store)

	// Check model path exists and is writable
	if info, err := os.Stat(path); err != nil {
		return nil, err
	} else if !info.IsDir() {
		return nil, ErrBadParameter.With("not a directory:", path)
	}

	// Get a listing of the models
	store.path = path
	store.ext = ext
	if err := store.Rescan(); err != nil {
		return nil, err
	}

	// Create a client
	if client := whisper.NewClient(modelUrl); client == nil {
		return nil, ErrInternalAppError
	} else {
		store.client = client
	}

	// Return success
	return store, nil
}

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s *Store) MarshalJSON() ([]byte, error) {
	modelNames := func() []string {
		result := make([]string, len(s.models))
		for i, model := range s.models {
			result[i] = model.Id
		}
		return result
	}
	return json.Marshal(struct {
		Path   string   `json:"path"`
		Ext    string   `json:"ext,omitempty"`
		Models []string `json:"models"`
	}{
		Path:   s.path,
		Ext:    s.ext,
		Models: modelNames(),
	})
}

func (s *Store) String() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the models
func (s *Store) List() []*schema.Model {
	s.RLock()
	defer s.RUnlock()
	return s.models
}

// Rescan models directory
func (s *Store) Rescan() error {
	s.Lock()
	defer s.Unlock()
	if models, err := listModels(s.path, s.ext); err != nil {
		return err
	} else {
		s.models = models
	}
	return nil
}

// Return a model by its Id
func (s *Store) ById(id string) *schema.Model {
	s.RLock()
	defer s.RUnlock()

	for _, model := range s.models {
		if model.Id == id {
			return model
		}
	}
	return nil
}

// Return a model by path
func (s *Store) ByPath(path string) *schema.Model {
	s.RLock()
	defer s.RUnlock()
	for _, model := range s.models {
		if model.Path == path {
			return model
		}
	}
	return nil
}

// Delete a model by its Id
func (s *Store) Delete(id string) error {
	model := s.ById(id)
	if model == nil {
		return ErrNotFound.Withf("%q", id)
	}

	// Lock the store
	s.Lock()
	defer s.Unlock()

	// Delete the model
	path := filepath.Join(s.path, model.Path)
	if err := os.Remove(path); err != nil {
		return err
	}

	// Rescan the models directory
	if models, err := listModels(s.path, s.ext); err != nil {
		return err
	} else {
		s.models = models
	}

	// Return success
	return nil
}

// Download a model to the models directory. If the model already exists, it will be returned
// without downloading. The destination directory is relative to the models directory.
//
// A function can be provided to track the progress of the download. If no Content-Length is
// provided by the server, the total bytes will be unknown and is set to zero.
func (s *Store) Download(ctx context.Context, path string, fn func(curBytes, totalBytes uint64)) (*schema.Model, error) {
	var remote string
	if u, err := url.Parse(path); err == nil {
		if u.Scheme != "" {
			remote = u.String()
		}
		path = u.Path
	}

	// abspath should be contained within the models directory
	abspath := filepath.Clean(filepath.Join(s.path, path))
	if !strings.HasPrefix(abspath, s.path) {
		return nil, ErrBadParameter.With(path)
	}

	// Get the model by path relative to the models directory
	relpath, err := filepath.Rel(s.path, abspath)
	if err != nil {
		return nil, err
	}
	model := s.ByPath(relpath)
	if model != nil {
		return model, nil
	}

	// File extension should match the store extension
	if s.ext != "" && filepath.Ext(abspath) != s.ext {
		return nil, ErrBadParameter.Withf("Bad file extension: %q", filepath.Base(abspath))
	}

	// Create the destination directory if it's not empty
	absdir := filepath.Dir(abspath)
	if info, err := os.Stat(absdir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(absdir, 0755); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if !info.IsDir() {
		return nil, ErrBadParameter.With(path)
	}

	// Create the destination file
	f, err := os.Create(abspath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Download the model, with callback. If an error occurs, the model is deleted again
	if _, err := s.client.Get(ctx, &writer{Writer: f, fn: fn}, filepath.Base(abspath), whisper.WithRemote(remote)); err != nil {
		return nil, errors.Join(toError(err), os.Remove(f.Name()))
	}

	// Rescan the models directory
	if err := s.Rescan(); err != nil {
		return nil, err
	}

	// Get a model by path
	model = s.ByPath(relpath)
	if model == nil {
		return nil, ErrNotFound.With(relpath)
	}

	// Return success
	return model, nil
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Convert 404 errors to ErrNotFound
func toError(err error) error {
	if err == nil {
		return nil
	}
	switch err := err.(type) {
	case *whisper.HTTPError:
		if err.Code == http.StatusNotFound {
			return ErrNotFound.With(err.Message)
		}
	}
	return err
}

func listModels(path, ext string) ([]*schema.Model, error) {
	result := make([]*schema.Model, 0, 100)

	// Walk filesystem
	return result, fs.WalkDir(os.DirFS(path), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Ignore hidden files or files without a .bin extension
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		if ext != "" && filepath.Ext(d.Name()) != ext {
			return nil
		}

		// Ignore files we can't get information on
		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Ignore non-regular files
		if !d.Type().IsRegular() {
			return nil
		}

		// Ignore files less than 8MB
		if info.Size() < 8*1024*1024 {
			return nil
		}

		// Get model information
		model := new(schema.Model)
		model.Object = "model"
		model.Path = path
		model.Created = info.ModTime().Unix()

		// Generate an Id for the model
		model.Id = modelNameToId(filepath.Base(path))

		// Append to result
		result = append(result, model)

		// Continue walking
		return nil
	})
}

func modelNameToId(name string) string {
	// Lowercase the name, remove the extension
	name = strings.TrimSuffix(strings.ToLower(name), filepath.Ext(name))

	// We replace all non-alphanumeric characters with underscores
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r
		}
		if r >= '0' && r <= '9' {
			return r
		}
		if r == '.' || r == '-' {
			return r
		}
		return '_'
	}, name)
}
