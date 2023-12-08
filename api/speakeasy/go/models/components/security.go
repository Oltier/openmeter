// Code generated by Speakeasy (https://speakeasyapi.dev). DO NOT EDIT.

package components

type Security struct {
	PortalToken string `security:"scheme,type=http,subtype=bearer,name=Authorization"`
}

func (o *Security) GetPortalToken() string {
	if o == nil {
		return ""
	}
	return o.PortalToken
}
