package storage

import "github.com/goj/maccbrowser/profile"

type ProfileStorage interface {
	Load() error
	List() ([]profile.Profile, error)
	Map() (map[string]profile.Profile, error)
	Create(profile.Profile) (profile.Profile, error)
	Delete(profile.Profile) error
	Len() (int, error)
	Get(id string) (profile.Profile, error)
}
