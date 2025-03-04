// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package config

import (
	"sync"
)

// Ensure, that IConfigMock does implement IConfig.
// If this is not the case, regenerate this file with moq.
var _ IConfig = &IConfigMock{}

// IConfigMock is a mock implementation of IConfig.
//
//     func TestSomethingThatUsesIConfig(t *testing.T) {
//
//         // make and configure a mocked IConfig
//         mockedIConfig := &IConfigMock{
//             LoadFunc: func() (*Config, error) {
// 	               panic("mock out the Load method")
//             },
//             LocationFunc: func() (string, error) {
// 	               panic("mock out the Location method")
//             },
//             RemoveFunc: func() error {
// 	               panic("mock out the Remove method")
//             },
//             SaveFunc: func(config *Config) error {
// 	               panic("mock out the Save method")
//             },
//         }
//
//         // use mockedIConfig in code that requires IConfig
//         // and then make assertions.
//
//     }
type IConfigMock struct {
	// LoadFunc mocks the Load method.
	LoadFunc func() (*Config, error)

	// LocationFunc mocks the Location method.
	LocationFunc func() (string, error)

	// RemoveFunc mocks the Remove method.
	RemoveFunc func() error

	// SaveFunc mocks the Save method.
	SaveFunc func(config *Config) error

	// calls tracks calls to the methods.
	calls struct {
		// Load holds details about calls to the Load method.
		Load []struct {
		}
		// Location holds details about calls to the Location method.
		Location []struct {
		}
		// Remove holds details about calls to the Remove method.
		Remove []struct {
		}
		// Save holds details about calls to the Save method.
		Save []struct {
			// Config is the config argument value.
			Config *Config
		}
	}
	lockLoad     sync.RWMutex
	lockLocation sync.RWMutex
	lockRemove   sync.RWMutex
	lockSave     sync.RWMutex
}

// Load calls LoadFunc.
func (mock *IConfigMock) Load() (*Config, error) {
	if mock.LoadFunc == nil {
		panic("IConfigMock.LoadFunc: method is nil but IConfig.Load was just called")
	}
	callInfo := struct {
	}{}
	mock.lockLoad.Lock()
	mock.calls.Load = append(mock.calls.Load, callInfo)
	mock.lockLoad.Unlock()
	return mock.LoadFunc()
}

// LoadCalls gets all the calls that were made to Load.
// Check the length with:
//     len(mockedIConfig.LoadCalls())
func (mock *IConfigMock) LoadCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockLoad.RLock()
	calls = mock.calls.Load
	mock.lockLoad.RUnlock()
	return calls
}

// Location calls LocationFunc.
func (mock *IConfigMock) Location() (string, error) {
	if mock.LocationFunc == nil {
		panic("IConfigMock.LocationFunc: method is nil but IConfig.Location was just called")
	}
	callInfo := struct {
	}{}
	mock.lockLocation.Lock()
	mock.calls.Location = append(mock.calls.Location, callInfo)
	mock.lockLocation.Unlock()
	return mock.LocationFunc()
}

// LocationCalls gets all the calls that were made to Location.
// Check the length with:
//     len(mockedIConfig.LocationCalls())
func (mock *IConfigMock) LocationCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockLocation.RLock()
	calls = mock.calls.Location
	mock.lockLocation.RUnlock()
	return calls
}

// Remove calls RemoveFunc.
func (mock *IConfigMock) Remove() error {
	if mock.RemoveFunc == nil {
		panic("IConfigMock.RemoveFunc: method is nil but IConfig.Remove was just called")
	}
	callInfo := struct {
	}{}
	mock.lockRemove.Lock()
	mock.calls.Remove = append(mock.calls.Remove, callInfo)
	mock.lockRemove.Unlock()
	return mock.RemoveFunc()
}

// RemoveCalls gets all the calls that were made to Remove.
// Check the length with:
//     len(mockedIConfig.RemoveCalls())
func (mock *IConfigMock) RemoveCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockRemove.RLock()
	calls = mock.calls.Remove
	mock.lockRemove.RUnlock()
	return calls
}

// Save calls SaveFunc.
func (mock *IConfigMock) Save(config *Config) error {
	if mock.SaveFunc == nil {
		panic("IConfigMock.SaveFunc: method is nil but IConfig.Save was just called")
	}
	callInfo := struct {
		Config *Config
	}{
		Config: config,
	}
	mock.lockSave.Lock()
	mock.calls.Save = append(mock.calls.Save, callInfo)
	mock.lockSave.Unlock()
	return mock.SaveFunc(config)
}

// SaveCalls gets all the calls that were made to Save.
// Check the length with:
//     len(mockedIConfig.SaveCalls())
func (mock *IConfigMock) SaveCalls() []struct {
	Config *Config
} {
	var calls []struct {
		Config *Config
	}
	mock.lockSave.RLock()
	calls = mock.calls.Save
	mock.lockSave.RUnlock()
	return calls
}
