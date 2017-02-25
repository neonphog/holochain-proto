// Copyright (C) 2013-2017, The MetaCurrency Project (Eric Harris-Braun, Arthur Brock, et. al.)
// Use of this source code is governed by GPLv3 found in the LICENSE file
//----------------------------------------------------------------------------------------
// Service implements functions and data that provide Holochain services

package holochain

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
)

// System settings, directory, and file names
const (
	DefaultDirectoryName string = ".holochain"  // Directory for storing config data
	DNAFileName          string = "dna"         // Definition of the Holochain
	ConfigFileName       string = "config"      // Settings of the Holochain
	SysFileName          string = "system.conf" // Server & System settings
	AgentFileName        string = "agent.txt"   // User ID info
	PrivKeyFileName      string = "priv.key"    // Signing key - private
	StoreFileName        string = "chain"       // Filename for local data store

	DefaultPort = 6283
)

// ServiceConfig holds the service settings
type ServiceConfig struct {
	DefaultPeerModeAuthor  bool
	DefaultPeerModeDHTNode bool
}

// Holochain service data structure
type Service struct {
	Settings     ServiceConfig
	DefaultAgent Agent
	Path         string
}

//IsInitialized checks a path for a correctly set up .holochain directory
func IsInitialized(root string) bool {
	return dirExists(root) && fileExists(root+"/"+SysFileName) && fileExists(root+"/"+AgentFileName)
}

//Init initializes service defaults including a signing key pair for an agent
func Init(root string, agent AgentID) (service *Service, err error) {
	err = os.MkdirAll(root, os.ModePerm)
	if err != nil {
		return
	}
	s := Service{
		Settings: ServiceConfig{
			DefaultPeerModeDHTNode: true,
			DefaultPeerModeAuthor:  true,
		},
		Path: root,
	}

	err = writeToml(root, SysFileName, s.Settings, false)
	if err != nil {
		return
	}

	a, err := NewAgent(IPFS, agent)
	if err != nil {
		return
	}
	err = SaveAgent(root, a)
	if err != nil {
		return
	}

	s.DefaultAgent = a

	service = &s
	return
}

func LoadService(path string) (service *Service, err error) {
	agent, err := LoadAgent(path)
	if err != nil {
		return
	}
	s := Service{
		Path:         path,
		DefaultAgent: agent,
	}

	_, err = toml.DecodeFile(path+"/"+SysFileName, &s.Settings)
	if err != nil {
		return
	}

	service = &s
	return
}

// ConfiguredChains returns a list of the configured chains for the given service
func (s *Service) ConfiguredChains() (chains map[string]*Holochain, err error) {
	files, err := ioutil.ReadDir(s.Path)
	if err != nil {
		return
	}
	chains = make(map[string]*Holochain)
	for _, f := range files {
		if f.IsDir() {
			h, err := s.Load(f.Name())
			if err == nil {
				chains[f.Name()] = h
			}
		}
	}
	return
}
