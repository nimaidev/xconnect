package main

// BACnet Application Tag constants
const (
	APP_TAG_NULL             = 0x00
	APP_TAG_BOOLEAN          = 0x10
	APP_TAG_UNSIGNED_INT     = 0x20
	APP_TAG_SIGNED_INT       = 0x30
	APP_TAG_REAL             = 0x40
	APP_TAG_DOUBLE           = 0x50
	APP_TAG_OCTET_STRING     = 0x60
	APP_TAG_CHARACTER_STRING = 0x70
	APP_TAG_BIT_STRING       = 0x80
	APP_TAG_ENUMERATED       = 0x90
	APP_TAG_DATE             = 0xA0
	APP_TAG_TIME             = 0xB0
	APP_TAG_BACNET_OBJECT_ID = 0xC0
)

// Request Types
const (
	SERV_CONFIRM_REQ    = 0x00
	SERV_UN_CONFIRM_REQ = 0x01
)

// Service command Constants
const (
	CNCTX_CMD_WHO_IS = 0x08
)
