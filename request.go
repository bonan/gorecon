package gorecon

type request struct {
	Length      uint8
	ControlByte byte
	Data        []byte
}

func parseRequest(data []byte) (*request, error) {
	req := &request{
		Length:      data[0],
		ControlByte: data[1],
	}
	len := uint8(req.Length)
	req.Data = data[2:len]
	return req, nil
}

func (r *request) checksum() byte {
	dataLen := byte(r.Length)
	checksum := dataLen + 0x01 + r.ControlByte
	for _, v := range r.Data {
		checksum += v
	}
	checksum = (((checksum ^ 0xFF) + 0x01) & 0xFF) + 0x01
	return checksum
}

func (r *request) byteArray() []byte {
	r.Length = uint8(len(r.Data) + 2)
	ret := []byte{
		byte(r.Length),
		r.ControlByte,
	}
	ret = append(ret, r.Data...)
	ret = append(ret, r.checksum())
	for len(ret) < 8 {
		ret = append(ret, 0x00)
	}
	return ret
}
