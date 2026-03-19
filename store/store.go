package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/johngillam/arena/model"
)

// Load reads the device list from a JSON file. Returns an empty slice if the file doesn't exist.
func Load(path string) ([]model.Device, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []model.Device{}, nil
		}
		return nil, fmt.Errorf("reading store: %w", err)
	}

	var devices []model.Device
	if err := json.Unmarshal(data, &devices); err != nil {
		return nil, fmt.Errorf("parsing store: %w", err)
	}
	return devices, nil
}

// Save writes the device list to a JSON file.
func Save(path string, devices []model.Device) error {
	data, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling store: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing store: %w", err)
	}
	return nil
}
