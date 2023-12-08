// Code generated by Speakeasy (https://speakeasyapi.dev). DO NOT EDIT.

package components

type Meter struct {
	// A unique identifier for the meter.
	ID *string `json:"id,omitempty"`
	// A unique identifier for the meter.
	Slug string `json:"slug"`
	// A description of the meter.
	Description *string `json:"description,omitempty"`
	// The aggregation type to use for the meter.
	Aggregation MeterAggregation `json:"aggregation"`
	// Aggregation window size.
	WindowSize WindowSize `json:"windowSize"`
	// The event type to aggregate.
	EventType string `json:"eventType"`
	// JSONPath expression to extract the value from the event data.
	ValueProperty *string `json:"valueProperty,omitempty"`
	// Named JSONPath expressions to extract the group by values from the event data.
	GroupBy map[string]string `json:"groupBy,omitempty"`
}

func (o *Meter) GetID() *string {
	if o == nil {
		return nil
	}
	return o.ID
}

func (o *Meter) GetSlug() string {
	if o == nil {
		return ""
	}
	return o.Slug
}

func (o *Meter) GetDescription() *string {
	if o == nil {
		return nil
	}
	return o.Description
}

func (o *Meter) GetAggregation() MeterAggregation {
	if o == nil {
		return MeterAggregation("")
	}
	return o.Aggregation
}

func (o *Meter) GetWindowSize() WindowSize {
	if o == nil {
		return WindowSize("")
	}
	return o.WindowSize
}

func (o *Meter) GetEventType() string {
	if o == nil {
		return ""
	}
	return o.EventType
}

func (o *Meter) GetValueProperty() *string {
	if o == nil {
		return nil
	}
	return o.ValueProperty
}

func (o *Meter) GetGroupBy() map[string]string {
	if o == nil {
		return nil
	}
	return o.GroupBy
}
