package driver

import (
	"context"
	"fmt"
	"github.com/thingio/edge-device-std/errors"
	"github.com/thingio/edge-device-std/logger"
	"github.com/thingio/edge-device-std/models"
	"math/rand"
	"time"
)

type Method = func(ins map[string]*models.DeviceData) (outs map[string]*models.DeviceData, err error)
type MethodName = string
type MethodInputName = string
type MethodOutputName = string

const (
	MethodNameIntn    = "Intn"
	MethodInputNameN  = "n"
	MethodOutputNameV = "v"

	SubscribeInterval = 15 * time.Second
)

var supportedMethods = map[MethodName]Method{
	MethodNameIntn: intn,
}

func NewRandomNumberTwin(product *models.Product, device *models.Device) (models.DeviceTwin, error) {
	twin := &randomNumberTwin{
		product: product,
		device:  device,
	}
	return twin, nil
}

type randomNumberTwin struct {
	product *models.Product
	device  *models.Device

	properties map[models.ProductPropertyID]*models.ProductProperty // for property's reading and writing
	events     map[models.ProductEventID]*models.ProductEvent       // for event's subscribing
	methods    map[models.ProductMethodID]*models.ProductMethod     // for method's calling

	lg *logger.Logger

	ctx    context.Context
	cancel context.CancelFunc
}

func (r *randomNumberTwin) Initialize(lg *logger.Logger) error {
	r.lg = lg

	r.properties = make(map[models.ProductPropertyID]*models.ProductProperty)
	for _, property := range r.product.Properties {
		r.properties[property.Id] = property
	}
	r.events = make(map[models.ProductEventID]*models.ProductEvent)
	for _, event := range r.product.Events {
		r.events[event.Id] = event
	}
	r.methods = make(map[models.ProductMethodID]*models.ProductMethod)
	for _, method := range r.product.Methods {
		if _, ok := supportedMethods[method.Id]; !ok {
			return errors.NewCommonEdgeError(errors.DeviceTwin, fmt.Sprintf("unsupported method: %s", method.Id), nil)
		}
		r.methods[method.Id] = method
	}

	return nil
}

func (r *randomNumberTwin) Start(ctx context.Context) error {
	r.ctx, r.cancel = context.WithCancel(ctx)

	// aim to simulate unexpected exceptions
	n := rand.NewSource(time.Now().UnixNano()).Int63()
	switch n % 100 {
	case 0:
		return errors.DeviceTwin.Error("simulation error")
	default:
		return nil
	}
}

func (r *randomNumberTwin) Stop(force bool) error {
	r.cancel()

	return nil
}

func (r *randomNumberTwin) HealthCheck() (*models.DeviceStatus, error) {
	// aim to simulate unexpected exceptions
	n := rand.NewSource(time.Now().UnixNano()).Int63()
	switch n % 10 {
	case 0:
		return &models.DeviceStatus{Device: r.device, State: models.DeviceStateException}, nil
	default:
		return &models.DeviceStatus{Device: r.device, State: models.DeviceStateConnected}, nil
	}
}

func (r *randomNumberTwin) Read(propertyID models.ProductPropertyID) (map[models.ProductPropertyID]*models.DeviceData, error) {
	values := make(map[string]*models.DeviceData)
	if propertyID == models.DeviceDataMultiPropsID {
		for _, property := range r.properties {
			values[property.Id] = randnum(propertyID, property.FieldType)
		}
	} else {
		property, ok := r.properties[propertyID]
		if !ok {
			return nil, errors.NewCommonEdgeError(errors.BadRequest, fmt.Sprintf("undefined property: %s", property.Id), nil)
		}
		values[propertyID] = randnum(property.Id, property.FieldType)
	}
	return values, nil
}

func (r *randomNumberTwin) Write(propertyID models.ProductPropertyID, values map[models.ProductPropertyID]*models.DeviceData) error {
	return errors.NewCommonEdgeError(errors.MethodNotAllowed, fmt.Sprintf("the product[%s] doesn't support Write", r.product.ID), nil)
}

func (r *randomNumberTwin) Subscribe(eventID models.ProductEventID, bus chan<- *models.DeviceDataWrapper) error {
	event, ok := r.events[eventID]
	if !ok {
		return errors.NewCommonEdgeError(errors.BadRequest, fmt.Sprintf("undefined event: %s", eventID), nil)
	}

	go func() {
		ticker := time.NewTicker(SubscribeInterval)
		for {
			select {
			case <-ticker.C:
				props := make(map[models.ProductPropertyID]*models.DeviceData)
				for _, out := range event.Outs {
					props[out.Id] = randnum(out.Id, out.FieldType)
				}
				bus <- &models.DeviceDataWrapper{
					ProductID:  r.product.ID,
					DeviceID:   r.device.ID,
					FuncID:     eventID,
					Properties: props,
				}
			case <-r.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
	return nil
}

func (r *randomNumberTwin) Call(methodID models.ProductMethodID, ins map[models.ProductPropertyID]*models.DeviceData) (
	outs map[models.ProductPropertyID]*models.DeviceData, err error) {
	outs, err = supportedMethods[methodID](ins)
	if err != nil {
		return nil, err
	}
	return outs, nil
}

// randnum simulates reading property from the device.
func randnum(propertyID, valueType string) *models.DeviceData {
	var value interface{}
	switch valueType {
	case models.PropertyValueTypeFloat:
		value = rand.Float64()
	default:
		value = rand.Intn(100)
	}
	dd, _ := models.NewDeviceData(propertyID, valueType, value)
	return dd
}

// ints simulates calling method from the device.
func intn(ins map[string]*models.DeviceData) (outs map[string]*models.DeviceData, err error) {
	inputN, ok := ins[MethodInputNameN]
	if !ok {
		return nil, errors.NewCommonEdgeError(errors.BadRequest, fmt.Sprintf("the input %s is required", MethodInputNameN), nil)
	}
	n, err := inputN.IntValue()
	if err != nil {
		return nil, errors.BadRequest.Cause(err, "the input %s must be an integer", MethodInputNameN)
	}
	data, err := models.NewDeviceData(MethodOutputNameV, models.PropertyValueTypeInt, rand.Intn(int(n)))
	if err != nil {
		return nil, errors.DeviceTwin.Cause(err, "fail to construct the device data")
	}
	outs = map[string]*models.DeviceData{
		MethodOutputNameV: data,
	}
	return outs, nil
}
