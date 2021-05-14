package disk

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/goj/maccbrowser/profile"
)

const (
	profileRoot     = "./profile_storage"
	profileDescName = "profdesc.json"
	profileUUIDlen  = 36
)

var (
	ErrNotFound = errors.New("profile not found")
)

type StorageDesc struct {
	Name  string
	Proxy string
}

type Storage struct {
	root string

	profilesMx sync.Mutex
	profiles   map[string]profile.Profile
}

func NewStorage() (*Storage, error) {
	err := os.MkdirAll(profileRoot, 0644)
	if err != nil {
		return nil, fmt.Errorf("creating profile root: %w", err)
	}

	return &Storage{root: profileRoot}, nil
}

func (s *Storage) Load() error {
	dirs, err := os.ReadDir(s.root)
	if err != nil {
		return fmt.Errorf("read profile root: %w", err)
	}

	s.profilesMx.Lock()
	defer s.profilesMx.Unlock()

	s.profiles = make(map[string]profile.Profile, len(dirs))

	for _, d := range dirs {
		if len(d.Name()) != profileUUIDlen {
			continue
		}

		pf, err := os.Open(filepath.Join(s.root, d.Name(), profileDescName))
		if err != nil {
			return fmt.Errorf("read profile: %w", err)
		}

		var profDesc StorageDesc

		err = json.NewDecoder(pf).Decode(&profDesc)
		if err != nil {
			pf.Close()

			return fmt.Errorf("decode profile: %w", err)
		}

		pf.Close()

		s.profiles[d.Name()] = profile.Profile{
			GUID:  d.Name(),
			Name:  profDesc.Name,
			Proxy: profDesc.Proxy,
		}
	}

	return nil
}

func (s *Storage) List() ([]profile.Profile, error) {
	s.profilesMx.Lock()

	ret := make([]profile.Profile, 0, len(s.profiles))
	for _, value := range s.profiles {
		ret = append(ret, value)
	}

	s.profilesMx.Unlock()

	return ret, nil
}

func (s *Storage) Map() (map[string]profile.Profile, error) {
	s.profilesMx.Lock()

	ret := make(map[string]profile.Profile, len(s.profiles))
	for key, value := range s.profiles {
		ret[key] = value
	}

	s.profilesMx.Unlock()

	return ret, nil
}

func (s *Storage) Len() (int, error) {
	s.profilesMx.Lock()
	ret := len(s.profiles)
	s.profilesMx.Unlock()

	return ret, nil
}

func (s *Storage) Create(p profile.Profile) (profile.Profile, error) {
	if p.GUID == "" {
		profUUID, err := uuid.NewV4()
		if err != nil {
			return p, fmt.Errorf("generate profile uuid: %w", err)
		}

		p.GUID = profUUID.String()
	}

	s.profilesMx.Lock()
	s.profiles[p.GUID] = p
	s.profilesMx.Unlock()

	return p, nil
}

func (s *Storage) Delete(p profile.Profile) error {
	s.profilesMx.Lock()
	delete(s.profiles, p.GUID)
	s.profilesMx.Unlock()

	return nil
}

func (s *Storage) Get(id string) (profile.Profile, error) {
	s.profilesMx.Lock()
	ret, ok := s.profiles[id]
	s.profilesMx.Unlock()

	if !ok {
		return ret, ErrNotFound
	}

	return ret, nil
}
