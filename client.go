package main

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/brianmario/screenlogic-homekit/screenlogic"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/log"
)

// This acts as a wrapper for screenlogic.Gateway which handles reconnection, caching
// and thread-safety.
//
// It also provides helper methods to simplify getting at the data we need for HomeKit.
type Client struct {
	gateway          *screenlogic.Gateway
	requestMutex     sync.Mutex
	clientName       string
	reconnectRetries uint8
	cache            struct {
		defaultExpiry time.Duration
		poolStatus    struct {
			last     *screenlogic.PoolStatus
			deadline int64
		}
		controllerConfig struct {
			last     *screenlogic.ControllerConfiguration
			deadline int64
		}
	}
}

func NewConnectedClient(clientName string) (*Client, error) {
	client := &Client{
		clientName:       clientName,
		reconnectRetries: 1, // TODO: make this configurable?
	}

	client.cache.defaultExpiry = time.Minute * 1

	err := client.connectToGateway()
	if err != nil {
		return nil, err
	}

	return client, err
}

func (c *Client) connectToGateway() error {
	gateway, err := screenlogic.DiscoverGateway()
	if err != nil {
		return err
	}

	err = gateway.Connect()
	if err != nil {
		return err
	}

	err = gateway.Login(c.clientName)
	if err != nil {
		return err
	}

	c.gateway = gateway

	return nil
}

func (c *Client) GetAirTemperature() (float64, error) {
	status, err := c.getPoolStatus()
	if err != nil {
		panic(err)
	}

	return c.convertTempToHomeKit(status.AirTemp), nil
}

func (c *Client) GetGatewayName() string {
	return c.gateway.Name
}

func (c *Client) GetGatewayVersion() (string, error) {
	version, err := c.gateway.Version()
	if err != nil {
		panic(err)
	}

	return version, nil
}

func (c *Client) GetTemperatureDisplayUnits() int {
	config, err := c.getControllerConfig()
	if err != nil {
		panic(err)
	}

	if config.IsCelcius {
		return characteristic.TemperatureDisplayUnitsCelsius
	}

	return characteristic.TemperatureDisplayUnitsFahrenheit
}

func (c *Client) GetCurrentPoolTemp() float64 {
	status, err := c.getPoolStatus()
	if err != nil {
		panic(err)
	}

	info := status.PoolWater()

	return c.convertTempToHomeKit(info.CurrentTemp)
}

func (c *Client) GetCurrentSpaTemp() float64 {
	status, err := c.getPoolStatus()
	if err != nil {
		panic(err)
	}

	info := status.SpaWater()

	return c.convertTempToHomeKit(info.CurrentTemp)
}

func (c *Client) GetPoolHeaterActive() int {
	status, err := c.getPoolStatus()
	if err != nil {
		panic(err)
	}

	info := status.PoolWater()

	if screenlogic.HeatMode(info.HeatMode) == screenlogic.HeatModeOn {
		return characteristic.ActiveActive
	}

	return characteristic.ActiveInactive
}

func (c *Client) GetSpaHeaterActive() int {
	status, err := c.getPoolStatus()
	if err != nil {
		panic(err)
	}

	info := status.SpaWater()

	if screenlogic.HeatMode(info.HeatMode) == screenlogic.HeatModeOn {
		return characteristic.ActiveActive
	}

	return characteristic.ActiveInactive
}

func (c *Client) GetPoolCurrentHeatingState() int {
	status, err := c.getPoolStatus()
	if err != nil {
		panic(err)
	}

	info := status.PoolWater()

	if info.HeaterStatus == 1 {
		return characteristic.CurrentHeaterCoolerStateHeating
	}

	return characteristic.CurrentHeaterCoolerStateInactive
}

func (c *Client) GetSpaCurrentHeatingState() int {
	status, err := c.getPoolStatus()
	if err != nil {
		panic(err)
	}

	info := status.SpaWater()

	if info.HeaterStatus == 1 {
		return characteristic.CurrentHeaterCoolerStateHeating
	}

	return characteristic.CurrentHeaterCoolerStateInactive
}

func (c *Client) GetPoolTargetHeatingState() int {
	status, err := c.getPoolStatus()
	if err != nil {
		panic(err)
	}

	info := status.PoolWater()

	switch screenlogic.HeatMode(info.HeatMode) {
	case screenlogic.HeatModeOff:
		return characteristic.TargetHeaterCoolerStateAuto
	case screenlogic.HeatModeOn:
		return characteristic.TargetHeaterCoolerStateHeat
	case screenlogic.HeatModeSolarOnly:
		return characteristic.TargetHeaterCoolerStateHeat
	case screenlogic.HeatModeSolarPreferred:
		return characteristic.TargetHeaterCoolerStateHeat
	default:
		return characteristic.TargetHeaterCoolerStateAuto
	}
}

func (c *Client) GetSpaTargetHeatingState() int {
	status, err := c.getPoolStatus()
	if err != nil {
		panic(err)
	}

	info := status.SpaWater()

	switch screenlogic.HeatMode(info.HeatMode) {
	case screenlogic.HeatModeOff:
		return characteristic.TargetHeaterCoolerStateAuto
	case screenlogic.HeatModeOn:
		return characteristic.TargetHeaterCoolerStateHeat
	case screenlogic.HeatModeSolarOnly:
		return characteristic.TargetHeaterCoolerStateHeat
	case screenlogic.HeatModeSolarPreferred:
		return characteristic.TargetHeaterCoolerStateHeat
	default:
		return characteristic.TargetHeaterCoolerStateAuto
	}
}

func (c *Client) GetPoolHeatingThresholdTemp() float64 {
	status, err := c.getPoolStatus()
	if err != nil {
		panic(err)
	}

	info := status.PoolWater()

	return c.convertTempToHomeKit(info.HeatSetPoint)
}

func (c *Client) GetSpaHeatingThresholdTemp() float64 {
	status, err := c.getPoolStatus()
	if err != nil {
		panic(err)
	}

	info := status.SpaWater()

	return c.convertTempToHomeKit(info.HeatSetPoint)
}

func (c *Client) getControllerConfig() (*screenlogic.ControllerConfiguration, error) {
	c.requestMutex.Lock()
	defer c.requestMutex.Unlock()

	if c.cache.controllerConfig.last != nil && time.Now().UnixNano() < c.cache.controllerConfig.deadline {
		return c.cache.controllerConfig.last, nil
	}

	retriesLeft := c.reconnectRetries

	var err error

retry:
	c.cache.controllerConfig.last, err = c.gateway.ControllerConfig()
	if err != nil {
		_, ok := err.(net.Error)

		if (ok || err == io.EOF) && retriesLeft > 0 {
			retriesLeft--

			// connection most likely dropped, let's reconnect
			log.Debug.Printf("getControllerConfig() - reconnect client attemp %d\n", c.reconnectRetries-retriesLeft)

			err = c.gateway.Reconnect()
			if err != nil {
				// if we still get an error here, give up
				panic(err)
			}

			goto retry
		}

		// some other error happened, give up
		return nil, err
	}

	c.cache.controllerConfig.deadline = time.Now().Add(c.cache.defaultExpiry).UnixNano()

	return c.cache.controllerConfig.last, nil
}

func (c *Client) getPoolStatus() (*screenlogic.PoolStatus, error) {
	c.requestMutex.Lock()
	defer c.requestMutex.Unlock()

	if c.cache.poolStatus.last != nil && time.Now().UnixNano() < c.cache.poolStatus.deadline {
		return c.cache.poolStatus.last, nil
	}

	retriesLeft := c.reconnectRetries

	var err error

retry:
	c.cache.poolStatus.last, err = c.gateway.PoolStatus()
	if err != nil {
		_, ok := err.(net.Error)

		if (ok || err == io.EOF) && retriesLeft > 0 {
			retriesLeft--

			// connection most likely dropped, let's reconnect
			log.Debug.Printf("getPoolStatus() - reconnect client attemp %d\n", c.reconnectRetries-retriesLeft)

			err = c.gateway.Reconnect()
			if err != nil {
				// if we still get an error here, give up
				panic(err)
			}

			goto retry
		}

		// some other error happened, give up
		return nil, err
	}

	c.cache.poolStatus.deadline = time.Now().Add(c.cache.defaultExpiry).UnixNano()

	return c.cache.poolStatus.last, nil
}

func (c *Client) celsiusToFahrenheit(celsius uint32) float64 {
	return float64(celsius*9/5 + 32)
}

func (c *Client) fahrenheitToCelsius(fahrenheit uint32) float64 {
	return float64((fahrenheit - 32) * 5 / 9)
}

// HomeKit always wants values to be in celsius, but the pool controller may be configured
// for Fahrenheit. This should always give us what we want.
func (c *Client) convertTempToHomeKit(poolTemp uint32) float64 {
	units := c.GetTemperatureDisplayUnits()

	switch units {
	case characteristic.TemperatureDisplayUnitsCelsius:
		return float64(poolTemp)
	case characteristic.TemperatureDisplayUnitsFahrenheit:
		return c.fahrenheitToCelsius(poolTemp)
	default:
		return c.fahrenheitToCelsius(poolTemp)
	}
}
