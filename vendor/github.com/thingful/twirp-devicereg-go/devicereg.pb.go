// Code generated by protoc-gen-go. DO NOT EDIT.
// source: devicereg.proto

/*
Package devicereg is a generated protocol buffer package.

It is generated from these files:
	devicereg.proto

It has these top-level messages:
	ClaimDeviceRequest
	ClaimDeviceResponse
	RevokeDeviceRequest
	RevokeDeviceResponse
*/
package devicereg

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// An enumeration which allows us to express whether the device will be
// located indoors or outdoors when deployed.
type ClaimDeviceRequest_Disposition int32

const (
	ClaimDeviceRequest_INDOOR  ClaimDeviceRequest_Disposition = 0
	ClaimDeviceRequest_OUTDOOR ClaimDeviceRequest_Disposition = 1
)

var ClaimDeviceRequest_Disposition_name = map[int32]string{
	0: "INDOOR",
	1: "OUTDOOR",
}
var ClaimDeviceRequest_Disposition_value = map[string]int32{
	"INDOOR":  0,
	"OUTDOOR": 1,
}

func (x ClaimDeviceRequest_Disposition) String() string {
	return proto.EnumName(ClaimDeviceRequest_Disposition_name, int32(x))
}
func (ClaimDeviceRequest_Disposition) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor0, []int{0, 0}
}

// ClaimDeviceRequest is the message we send in order to initially claim that a
// specific user owns a device. This message contains the device token (which
// identifies the device), the individual's DECODE user id as well as some
// metadata about the device. Currently this is just the lat/long location of
// the device, and an enumarated value describing whether the claimed device is
// situated indoors or outdoors.
// As a result of this message the device registration service creates a key
// pair for the device as well as a key pair for the user.
type ClaimDeviceRequest struct {
	// The unique identifier for the device. Note this isn't a hardware identifier
	// as the same physical device may go to multiple recipients, rather it
	// represents the logical ID of the device as currently claimed. This comes
	// from SmartCitizen's onboarding process ultimately. This is a required field
	DeviceToken string `protobuf:"bytes,1,opt,name=device_token,json=deviceToken" json:"device_token,omitempty"`
	// A unique identifier for the user which should come from the DECODE wallet
	// ultimately. This is a required field.
	UserUid string `protobuf:"bytes,2,opt,name=user_uid,json=userUid" json:"user_uid,omitempty"`
	// The location of the device to be claimed. This is a required field.
	Location *ClaimDeviceRequest_Location `protobuf:"bytes,3,opt,name=location" json:"location,omitempty"`
	// The specific disposition of the device, i.e. is this instance indoors or
	// outdoors. If not specified the default value is INDOOR.
	Disposition ClaimDeviceRequest_Disposition `protobuf:"varint,4,opt,name=disposition,enum=devicereg.ClaimDeviceRequest_Disposition" json:"disposition,omitempty"`
	// The address of the MQTT broker to which the specified device is configured
	// to publish data. This is a required field.
	Broker string `protobuf:"bytes,5,opt,name=broker" json:"broker,omitempty"`
}

func (m *ClaimDeviceRequest) Reset()                    { *m = ClaimDeviceRequest{} }
func (m *ClaimDeviceRequest) String() string            { return proto.CompactTextString(m) }
func (*ClaimDeviceRequest) ProtoMessage()               {}
func (*ClaimDeviceRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *ClaimDeviceRequest) GetDeviceToken() string {
	if m != nil {
		return m.DeviceToken
	}
	return ""
}

func (m *ClaimDeviceRequest) GetUserUid() string {
	if m != nil {
		return m.UserUid
	}
	return ""
}

func (m *ClaimDeviceRequest) GetLocation() *ClaimDeviceRequest_Location {
	if m != nil {
		return m.Location
	}
	return nil
}

func (m *ClaimDeviceRequest) GetDisposition() ClaimDeviceRequest_Disposition {
	if m != nil {
		return m.Disposition
	}
	return ClaimDeviceRequest_INDOOR
}

func (m *ClaimDeviceRequest) GetBroker() string {
	if m != nil {
		return m.Broker
	}
	return ""
}

// A nested type capturing the location of the device expressed via decimal
// long/lat pair.
type ClaimDeviceRequest_Location struct {
	// The longitude expressed as a decimal. This is a required field.
	Longitude float64 `protobuf:"fixed64,1,opt,name=longitude" json:"longitude,omitempty"`
	// The latitude expressed as a decimal. This is a required field.
	Latitude float64 `protobuf:"fixed64,2,opt,name=latitude" json:"latitude,omitempty"`
}

func (m *ClaimDeviceRequest_Location) Reset()                    { *m = ClaimDeviceRequest_Location{} }
func (m *ClaimDeviceRequest_Location) String() string            { return proto.CompactTextString(m) }
func (*ClaimDeviceRequest_Location) ProtoMessage()               {}
func (*ClaimDeviceRequest_Location) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

func (m *ClaimDeviceRequest_Location) GetLongitude() float64 {
	if m != nil {
		return m.Longitude
	}
	return 0
}

func (m *ClaimDeviceRequest_Location) GetLatitude() float64 {
	if m != nil {
		return m.Latitude
	}
	return 0
}

// ClaimDeviceResponse is the message returned after successfully claiming a
// device. We return here a key pair for the user, as well as a public key for
// the device. The corresponding private key is used within the stream encoder
// in order to encrypt data for the device.
type ClaimDeviceResponse struct {
	// The private part of a key pair for the individual user.
	UserPrivateKey string `protobuf:"bytes,1,opt,name=user_private_key,json=userPrivateKey" json:"user_private_key,omitempty"`
	// The public part of a key pair representing the individual user.
	UserPublicKey string `protobuf:"bytes,2,opt,name=user_public_key,json=userPublicKey" json:"user_public_key,omitempty"`
	// The public key for the device (TODO - is this useful for any reason?)
	DevicePublicKey string `protobuf:"bytes,3,opt,name=device_public_key,json=devicePublicKey" json:"device_public_key,omitempty"`
}

func (m *ClaimDeviceResponse) Reset()                    { *m = ClaimDeviceResponse{} }
func (m *ClaimDeviceResponse) String() string            { return proto.CompactTextString(m) }
func (*ClaimDeviceResponse) ProtoMessage()               {}
func (*ClaimDeviceResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ClaimDeviceResponse) GetUserPrivateKey() string {
	if m != nil {
		return m.UserPrivateKey
	}
	return ""
}

func (m *ClaimDeviceResponse) GetUserPublicKey() string {
	if m != nil {
		return m.UserPublicKey
	}
	return ""
}

func (m *ClaimDeviceResponse) GetDevicePublicKey() string {
	if m != nil {
		return m.DevicePublicKey
	}
	return ""
}

// RevokeDeviceRequest is a message sent to the registration service by which a
// user can revoke a previous claim on a device. This should result in all
// configuration for the device being deleted from registration services store,
// as well removing any stream encoding configurations.
type RevokeDeviceRequest struct {
	// The unique token identifying the device.
	DeviceToken string `protobuf:"bytes,1,opt,name=device_token,json=deviceToken" json:"device_token,omitempty"`
	// The user's public key, serving here just to prove that the user actually is
	// the entity that previously claimed the device.
	UserPublicKey string `protobuf:"bytes,2,opt,name=user_public_key,json=userPublicKey" json:"user_public_key,omitempty"`
}

func (m *RevokeDeviceRequest) Reset()                    { *m = RevokeDeviceRequest{} }
func (m *RevokeDeviceRequest) String() string            { return proto.CompactTextString(m) }
func (*RevokeDeviceRequest) ProtoMessage()               {}
func (*RevokeDeviceRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *RevokeDeviceRequest) GetDeviceToken() string {
	if m != nil {
		return m.DeviceToken
	}
	return ""
}

func (m *RevokeDeviceRequest) GetUserPublicKey() string {
	if m != nil {
		return m.UserPublicKey
	}
	return ""
}

// RevokeDeviceResponse is a placeholder response returned from a revoke
// request. Currently empty, but reserved for any fields identified for future
// iterations.
type RevokeDeviceResponse struct {
}

func (m *RevokeDeviceResponse) Reset()                    { *m = RevokeDeviceResponse{} }
func (m *RevokeDeviceResponse) String() string            { return proto.CompactTextString(m) }
func (*RevokeDeviceResponse) ProtoMessage()               {}
func (*RevokeDeviceResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func init() {
	proto.RegisterType((*ClaimDeviceRequest)(nil), "devicereg.ClaimDeviceRequest")
	proto.RegisterType((*ClaimDeviceRequest_Location)(nil), "devicereg.ClaimDeviceRequest.Location")
	proto.RegisterType((*ClaimDeviceResponse)(nil), "devicereg.ClaimDeviceResponse")
	proto.RegisterType((*RevokeDeviceRequest)(nil), "devicereg.RevokeDeviceRequest")
	proto.RegisterType((*RevokeDeviceResponse)(nil), "devicereg.RevokeDeviceResponse")
	proto.RegisterEnum("devicereg.ClaimDeviceRequest_Disposition", ClaimDeviceRequest_Disposition_name, ClaimDeviceRequest_Disposition_value)
}

func init() { proto.RegisterFile("devicereg.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 396 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x93, 0x4f, 0xee, 0xd2, 0x40,
	0x14, 0xc7, 0x1d, 0x50, 0x28, 0xaf, 0x08, 0xf8, 0x30, 0xa4, 0x36, 0xfe, 0xc1, 0x2e, 0x48, 0x75,
	0xc1, 0x02, 0x6f, 0x80, 0xdd, 0x18, 0x88, 0x35, 0x0d, 0x6c, 0xdc, 0x60, 0xa1, 0x13, 0x32, 0x69,
	0xed, 0xd4, 0x76, 0x4a, 0xc2, 0x39, 0x3c, 0x8a, 0xa7, 0xf2, 0x16, 0xa6, 0x33, 0xa5, 0x2d, 0x51,
	0x88, 0xbf, 0xe5, 0xfb, 0xbe, 0xcf, 0xfb, 0x3f, 0x03, 0xc3, 0x80, 0x9e, 0xd8, 0x81, 0xa6, 0xf4,
	0x38, 0x4f, 0x52, 0x2e, 0x38, 0xf6, 0x2a, 0xc1, 0xfa, 0xdd, 0x02, 0xfc, 0x18, 0xf9, 0xec, 0xbb,
	0x23, 0x25, 0x8f, 0xfe, 0xc8, 0x69, 0x26, 0xf0, 0x2d, 0xf4, 0x15, 0xb3, 0x13, 0x3c, 0xa4, 0xb1,
	0x41, 0xa6, 0xc4, 0xee, 0x79, 0xba, 0xd2, 0x36, 0x85, 0x84, 0x2f, 0x40, 0xcb, 0x33, 0x9a, 0xee,
	0x72, 0x16, 0x18, 0x2d, 0xe9, 0xee, 0x16, 0xf6, 0x96, 0x05, 0xb8, 0x04, 0x2d, 0xe2, 0x07, 0x5f,
	0x30, 0x1e, 0x1b, 0xed, 0x29, 0xb1, 0xf5, 0xc5, 0x6c, 0x5e, 0xf7, 0xf0, 0x77, 0xb9, 0xf9, 0xba,
	0xa4, 0xbd, 0x2a, 0x0e, 0x57, 0xa0, 0x07, 0x2c, 0x4b, 0x78, 0xc6, 0x64, 0x9a, 0xc7, 0x53, 0x62,
	0x0f, 0x16, 0xef, 0xee, 0xa7, 0x71, 0xea, 0x00, 0xaf, 0x19, 0x8d, 0x13, 0xe8, 0xec, 0x53, 0x1e,
	0xd2, 0xd4, 0x78, 0x22, 0x3b, 0x2d, 0x2d, 0xd3, 0x01, 0xed, 0x52, 0x1a, 0x5f, 0x42, 0x2f, 0xe2,
	0xf1, 0x91, 0x89, 0x3c, 0xa0, 0x72, 0x5e, 0xe2, 0xd5, 0x02, 0x9a, 0xa0, 0x45, 0xbe, 0x50, 0xce,
	0x96, 0x74, 0x56, 0xb6, 0x35, 0x03, 0xbd, 0x51, 0x19, 0x01, 0x3a, 0x9f, 0x3e, 0x3b, 0xae, 0xeb,
	0x8d, 0x1e, 0xa1, 0x0e, 0x5d, 0x77, 0xbb, 0x91, 0x06, 0xb1, 0x7e, 0x12, 0x18, 0x5f, 0x75, 0x9d,
	0x25, 0x3c, 0xce, 0x28, 0xda, 0x30, 0x92, 0x9b, 0x4c, 0x52, 0x76, 0xf2, 0x05, 0xdd, 0x85, 0xf4,
	0x5c, 0x2e, 0x7c, 0x50, 0xe8, 0x5f, 0x94, 0xbc, 0xa2, 0x67, 0x9c, 0xc1, 0x50, 0x91, 0xf9, 0x3e,
	0x62, 0x07, 0x09, 0xaa, 0xd5, 0x3f, 0x95, 0xa0, 0x54, 0x0b, 0xee, 0x3d, 0x3c, 0x2b, 0xcf, 0xd7,
	0x20, 0xdb, 0x92, 0x2c, 0x1f, 0x43, 0xc5, 0x5a, 0xdf, 0x60, 0xec, 0xd1, 0x13, 0x0f, 0xe9, 0x83,
	0x5f, 0xc0, 0x7f, 0x76, 0x63, 0x4d, 0xe0, 0xf9, 0x75, 0x05, 0x35, 0xf7, 0xe2, 0x17, 0x01, 0xbc,
	0x48, 0x47, 0x96, 0x89, 0x54, 0x1d, 0x62, 0x0d, 0x7a, 0x63, 0x4b, 0xf8, 0xea, 0xee, 0xcd, 0xcd,
	0xd7, 0xb7, 0xdc, 0xe5, 0x72, 0x5d, 0xe8, 0x37, 0x8b, 0x63, 0x93, 0xff, 0xc7, 0xdc, 0xe6, 0x9b,
	0x9b, 0x7e, 0x95, 0x70, 0xa9, 0x7f, 0xad, 0xbf, 0xcf, 0xbe, 0x23, 0x3f, 0xd4, 0x87, 0x3f, 0x01,
	0x00, 0x00, 0xff, 0xff, 0x83, 0x06, 0x8d, 0x23, 0x63, 0x03, 0x00, 0x00,
}
