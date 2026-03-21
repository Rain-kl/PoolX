package geoip

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/oschwald/maxminddb-golang"
)

var GeoIpUrl = "https://raw.githubusercontent.com/Loyalsoldier/geoip/release/GeoLite2-Country.mmdb"
var GeoIpFilePath = "./data/GeoLite2-Country.mmdb"

type GeoIpRecord struct {
	Country struct {
		ISOCode string            `maxminddb:"iso_code"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
}

type MaxMindGeoIPService struct {
	maxMindDBReader *maxminddb.Reader
	dbFilePath      string
	mu              sync.RWMutex
}

func (s *MaxMindGeoIPService) Name() string {
	return "MaxMind"
}

func NewMaxMindGeoIPService() (*MaxMindGeoIPService, error) {
	service := &MaxMindGeoIPService{
		dbFilePath: GeoIpFilePath,
	}

	if err := os.MkdirAll(filepath.Dir(service.dbFilePath), os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create data directory for MaxMind database: %w", err)
	}

	if _, err := os.Stat(service.dbFilePath); os.IsNotExist(err) {
		if err := service.UpdateDatabase(); err != nil {
			return nil, fmt.Errorf("failed to download initial MaxMind database: %w", err)
		}
	}

	if err := service.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize MaxMind database: %w", err)
	}

	return service, nil
}

func (s *MaxMindGeoIPService) initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.maxMindDBReader != nil {
		_ = s.maxMindDBReader.Close()
		s.maxMindDBReader = nil
	}

	reader, err := maxminddb.Open(s.dbFilePath)
	if err != nil {
		return fmt.Errorf("error opening MaxMind database at %s: %w", s.dbFilePath, err)
	}
	s.maxMindDBReader = reader
	return nil
}

func (s *MaxMindGeoIPService) GetGeoInfo(ip net.IP) (*GeoInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.maxMindDBReader == nil {
		return nil, fmt.Errorf("MaxMind database is not initialized or failed to open")
	}
	if ip == nil {
		return nil, fmt.Errorf("IP address cannot be nil")
	}

	var record GeoIpRecord
	if err := s.maxMindDBReader.Lookup(ip, &record); err != nil {
		return nil, fmt.Errorf("error looking up IP %s in MaxMind database: %w", ip.String(), err)
	}

	geoInfo := &GeoInfo{
		ISOCode: record.Country.ISOCode,
		Name:    record.Country.Names["en"],
	}
	if geoInfo.Name == "" && geoInfo.ISOCode != "" {
		geoInfo.Name = geoInfo.ISOCode
	}

	return geoInfo, nil
}

func (s *MaxMindGeoIPService) UpdateDatabase() error {
	resp, err := http.Get(GeoIpUrl)
	if err != nil {
		return fmt.Errorf("failed to initiate MaxMind database download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download MaxMind database: HTTP status %s", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(s.dbFilePath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create data directory for MaxMind database update: %w", err)
	}

	tempPath := s.dbFilePath + ".download"
	out, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create MaxMind database file at %s: %w", tempPath, err)
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write MaxMind database file: %w", err)
	}
	if err = out.Close(); err != nil {
		return fmt.Errorf("failed to close MaxMind database file: %w", err)
	}
	if err = os.Rename(tempPath, s.dbFilePath); err != nil {
		return fmt.Errorf("failed to move MaxMind database file into place: %w", err)
	}

	return s.initialize()
}

func (s *MaxMindGeoIPService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.maxMindDBReader != nil {
		err := s.maxMindDBReader.Close()
		s.maxMindDBReader = nil
		if err != nil {
			return fmt.Errorf("error closing MaxMind database: %w", err)
		}
	}
	return nil
}
