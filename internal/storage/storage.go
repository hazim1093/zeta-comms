package storage

import (
	"os"

	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

type Data struct {
	Networks map[string]NetworkData `yaml:"networks"`
}

type NetworkData struct {
	LastProcessedProposalID string `yaml:"lastProcessedProposalId"`
}

type StorageService struct {
	config *config.Config
	log    *zerolog.Logger
}

func NewStorageService(cfg *config.Config, logger *zerolog.Logger) *StorageService {
	log := logger.With().Str("service", "storage").Logger()

	return &StorageService{
		config: cfg,
		log:    &log,
	}
}

func (s *StorageService) StoreLastProcessedProposalID(network string, proposalID string) error {
	// Load existing data
	data, err := loadYamlFile(s.config.Storage.Filename)
	if err != nil {
		// If file doesn't exist, create new data
		if os.IsNotExist(err) {
			data = Data{
				Networks: make(map[string]NetworkData),
			}
		} else {
			return err
		}
	}

	// Ensure Networks map is initialized
	if data.Networks == nil {
		data.Networks = make(map[string]NetworkData)
	}

	// Update or add the network's proposal ID
	data.Networks[network] = NetworkData{
		LastProcessedProposalID: proposalID,
	}

	return saveYamlFile(s.config.Storage.Filename, data)
}

func (s *StorageService) GetLastProcessedProposalID(network string) (string, error) {
	data, err := loadYamlFile(s.config.Storage.Filename)
	if err != nil {
		// If file doesn't exist, create new data
		if os.IsNotExist(err) {
			return "", nil // No data available for this network
		} else {
			return "", err
		}
	}

	if networkData, ok := data.Networks[network]; ok {
		return networkData.LastProcessedProposalID, nil
	}

	return "", nil
}

func saveYamlFile(filename string, data Data) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, yamlData, 0644)
}

func loadYamlFile(filename string) (Data, error) {
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return Data{}, err
	}

	var data Data

	err = yaml.Unmarshal(fileData, &data)
	if err != nil {
		return Data{}, err
	}

	return data, nil
}
