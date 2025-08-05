package main

// Helper to encode BACnet Object Identifier (ObjectType=8 for Device)
func encodeObjectID(objectType, instanceNumber uint32) uint32 {
	return (objectType << 22) | (instanceNumber & 0x3FFFFF)
}

// Encode BACnet Application Tag with data
func encodeBACnetApplicationTag(tagNumber byte, data []byte) []byte {
	var result []byte
	dataLen := len(data)

	if dataLen < 5 {
		// Length fits in 3 bits of tag byte
		tagByte := tagNumber | byte(dataLen)
		result = append(result, tagByte)
	} else {
		// Extended length encoding
		tagByte := tagNumber | 0x05 // 5 indicates extended length
		result = append(result, tagByte)
		if dataLen < 254 {
			result = append(result, byte(dataLen))
		} else {
			result = append(result, 254) // Extended length indicator
			result = append(result, byte(dataLen>>8), byte(dataLen&0xFF))
		}
	}

	result = append(result, data...)
	return result
}

// Encode unsigned integer
func encodeUnsignedInt(value uint32) []byte {
	if value < 256 {
		return []byte{byte(value)}
	} else if value < 65536 {
		return []byte{byte(value >> 8), byte(value & 0xFF)}
	} else if value < 16777216 {
		return []byte{byte(value >> 16), byte(value >> 8), byte(value & 0xFF)}
	} else {
		return []byte{byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value & 0xFF)}
	}
}

// Encode BACnet Object Identifier
func encodeBACnetObjectID(objectID uint32) []byte {
	return []byte{
		byte(objectID >> 24),
		byte(objectID >> 16),
		byte(objectID >> 8),
		byte(objectID & 0xFF),
	}
}
