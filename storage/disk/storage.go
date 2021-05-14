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
	profileDescName = "profdesc.json"
	profileUUIDlen  = 36
)

var ErrNotFound = errors.New("profile not found")

type Storage struct {
	root string

	profilesMx sync.Mutex
	profiles   map[string]profile.Profile
}

func NewStorage(root string) (*Storage, error) {
	if err := os.MkdirAll(root, 0644); err != nil {
		return nil, fmt.Errorf("creating profile root: %w", err)
	}

	return &Storage{root: root}, nil
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

		var profDesc profile.Profile

		err = json.NewDecoder(pf).Decode(&profDesc)
		if err != nil {
			pf.Close()

			return fmt.Errorf("decode profile: %w", err)
		}

		pf.Close()

		s.profiles[profDesc.GUID] = profDesc
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

	pdir := filepath.Join(s.root, p.GUID)

	err := os.MkdirAll(pdir, 0644)
	if err != nil {
		return p, fmt.Errorf("create dir: %w", err)
	}

	pfile, err := os.Open(filepath.Join(pdir, profileDescName))
	if err != nil {
		return p, fmt.Errorf("create desc file: %w", err)
	}

	err = json.NewEncoder(pfile).Encode(p)
	if err != nil {
		pfile.Close()

		return p, fmt.Errorf("encode profile desc: %w", err)
	}

	pfile.Close()

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
