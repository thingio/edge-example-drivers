package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/thingio/edge-device-std/errors"
	"github.com/thingio/edge-device-std/logger"
	"github.com/thingio/edge-device-std/models"
)

type Method = func(ins map[string]*models.DeviceData) (outs map[string]*models.DeviceData, err error)
type MethodName = string
type MethodInputName = string
type MethodOutputName = string

const (
	SubscribeInterval = 15 * time.Second
)

var supportedMethods = map[MethodName]Method{}

func NewStockTwin(product *models.Product, device *models.Device) (models.DeviceTwin, error) {
	if product == nil {
		return nil, errors.NewCommonEdgeError(errors.DeviceTwin, "the product cannot be nil", nil)
	}
	if device == nil {
		return nil, errors.NewCommonEdgeError(errors.DeviceTwin, "the device cannot be nil", nil)
	}

	twin := &stockDeviceTwin{
		product: product,
		device:  device,
	}
	return twin, nil
}

type stockDeviceTwin struct {
	product *models.Product
	device  *models.Device

	properties map[models.ProductPropertyID]*models.ProductProperty // for property's reading
	events     map[models.ProductEventID]*models.ProductEvent       // for event's subscribing
	methods    map[models.ProductMethodID]*models.ProductMethod     // for method's calling

	lg *logger.Logger

	ctx    context.Context
	cancel context.CancelFunc
}

func (r *stockDeviceTwin) Initialize(lg *logger.Logger) error {
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

func (r *stockDeviceTwin) Start(ctx context.Context) error {
	r.ctx, r.cancel = context.WithCancel(ctx)
	return nil
}

func (r *stockDeviceTwin) Stop(force bool) error {
	r.cancel()

	return nil
}

func (r *stockDeviceTwin) HealthCheck() (*models.DeviceStatus, error) {
	return &models.DeviceStatus{Device: r.device, State: models.DeviceStateConnected}, nil
}

func (r *stockDeviceTwin) Write(propertyID models.ProductPropertyID, values map[models.ProductPropertyID]*models.DeviceData) error {
	return errors.NewCommonEdgeError(errors.MethodNotAllowed, fmt.Sprintf("the product[%s] doesn't support Write", r.product.ID), nil)
}

func (r *stockDeviceTwin) Read(propertyID models.ProductPropertyID) (map[models.ProductEventID]*models.DeviceData, error) {
	quote, err := getQuote(r.device.ID)
	if err != nil {
		return nil, errors.NewCommonEdgeError(errors.Internal, fmt.Sprintf("failed to read the quote of the stock[%s]", r.device.ID), err)
	}

	values := make(map[string]*models.DeviceData)
	if propertyID == models.DeviceDataMultiPropsID {
		// multiple properties
		for _, property := range r.properties {
			values[property.Id] = quote[property.Id]
		}
	} else {
		// single property
		property, ok := r.properties[propertyID]
		if !ok {
			return nil, errors.NewCommonEdgeError(errors.BadRequest, fmt.Sprintf("undefined property: %s", property.Id), nil)
		}
		values[propertyID] = quote[propertyID]
	}

	return values, nil
}

func (r *stockDeviceTwin) Subscribe(eventID models.ProductEventID, bus chan<- *models.DeviceDataWrapper) error {
	_, ok := r.events[eventID]
	if !ok {
		return errors.NewCommonEdgeError(errors.BadRequest, fmt.Sprintf("undefined event: %s", eventID), nil)
	}

	go func() {
		ticker := time.NewTicker(SubscribeInterval)
		for {
			select {
			case <-ticker.C:
				quote, err := getQuote(r.device.ID)
				if err != nil {
					r.lg.WithError(err).Errorf("failed to read the quote of the stock[%s]", r.device.ID)
					return
				}
				bus <- &models.DeviceDataWrapper{
					ProductID:  r.product.ID,
					DeviceID:   r.device.ID,
					FuncID:     eventID,
					Properties: quote,
				}
			case <-r.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (r *stockDeviceTwin) Call(methodID models.ProductMethodID, ins map[models.ProductPropertyID]*models.DeviceData) (
	outs map[models.ProductPropertyID]*models.DeviceData, err error) {
	outs, err = supportedMethods[methodID](ins)
	if err != nil {
		return nil, err
	}
	return outs, nil
}

type XQData struct {
	Data             []map[string]interface{} `json:"data"`
	ErrorCode        int                      `json:"error_code"`
	ErrorDescription string                   `json:"error_description"`
}

func getQuote(deviceID string) (map[string]*models.DeviceData, error) {
	r, err := http.Get(fmt.Sprintf("https://stock.xueqiu.com/v5/stock/realtime/quotec.json?symbol=%s", deviceID))
	if err != nil {
		return nil, err
	}
	data := &XQData{}
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		return nil, err
	}
	result := make(map[string]*models.DeviceData)
	if len(data.Data) > 0 {
		for name, value := range data.Data[0] {
			result[name] = &models.DeviceData{
				Name:  name,
				Type:  fmt.Sprintf("%T", value),
				Value: value,
				Ts:    time.Now(),
			}
		}
	}
	return result, nil
}
