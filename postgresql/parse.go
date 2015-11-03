package postgresql

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"strconv"
)

func readInt32(c *PGConn) uint32 {
	var ret uint32 = uint32(c.buf[c.idx+0])
	ret <<= 8
	ret += uint32(c.buf[c.idx+1])
	ret <<= 8
	ret += uint32(c.buf[c.idx+2])
	ret <<= 8
	ret += uint32(c.buf[c.idx+3])
	c.idx += 4
	return ret
}

func readInt16(c *PGConn) uint16 {
	var ret uint16 = uint16(c.buf[c.idx+0])
	ret <<= 8
	ret += uint16(c.buf[c.idx+1])
	c.idx += 2
	return ret
}

func readInt8(c *PGConn) byte {
	c.idx += 1
	return c.buf[c.idx-1]
}

func readByteN(c *PGConn, len int) []byte {
	c.idx += len
	return c.buf[c.idx-len : c.idx]
}

func readString(c *PGConn) string {
	idx := bytes.IndexByte(c.buf[c.idx:], 0)
	c.idx += idx + 1
	return string(c.buf[c.idx-idx-1 : c.idx-1])
}

func getInt32FromByteSlice(bs []byte) uint32 {
	var ret uint32 = uint32(bs[0])
	ret <<= 8
	ret += uint32(bs[1])
	ret <<= 8
	ret += uint32(bs[2])
	ret <<= 8
	ret += uint32(bs[3])
	return ret
}

func getIntFromByteSlice(bs []byte, len int) int {
	ret, _ := strconv.Atoi(string(bs))
	return ret
}

func ReadMessage(c *PGConn) (message interface{}, err error) {
	header := make([]byte, 5)
	n, e := c.conn.Read(header)
	if e != nil {
		return nil, e
	}
	if n != 5 {
		return nil, fmt.Errorf("cannot read message header (%d)", n)
	}
	length := getInt32FromByteSlice(header[1:])
	if viper.GetBool("info") {
		fmt.Printf("type (%c)\n", header[0])
	}

	body := make([]byte, length-4)
	n, e = io.ReadFull(c.conn, body)
	if e != nil {
		return nil, e
	}
	if n != int(length)-4 {
		return nil, fmt.Errorf("cannot read message body (%d)", n)
	}

	c.buf = body
	c.idx = 0

	switch header[0] {
	case 'R':
		return readAuthenticationMessage(c, length)
	case 'S':
		name := readString(c)
		value := readString(c)
		return ParameterStatus{name, value}, nil
	case 'K':
		processId := readInt32(c)
		secretKey := readInt32(c)
		return BackendKeyData{processId, secretKey}, nil
	case 'Z':
		transactionStatus := readInt8(c)
		return ReadyForQuery{transactionStatus}, nil
	case 'C':
		commandTag := readString(c)
		return CommandComplete{commandTag}, nil
	case 'G': // CopyInResponse
		fallthrough
	case 'H': // CopyOutResponse, -Flush-
	case 'T':
		count := int(readInt16(c))
		fieldInfos := make([]FieldInfo, 0, count)
		for i := 0; i < count; i++ {
			name := readString(c)
			tableObjectId := readInt32(c)
			columnAttributeNum := readInt16(c)
			datatypeObjectId := readInt32(c)
			datatypeSize := readInt16(c)
			typeModifier := readInt32(c)
			formatCode := readInt16(c)
			fieldInfos = append(fieldInfos, FieldInfo{name, tableObjectId,
				columnAttributeNum, datatypeObjectId, datatypeSize,
				typeModifier, formatCode})
		}
		return RowDescription{fieldInfos}, nil
	case 'D':
		count := int(readInt16(c))
		valueInfos := make([]ValueInfo, 0, count)
		for i := 0; i < count; i++ {
			valueLen := readInt32(c)
			value := readByteN(c, int(valueLen))
			valueInfos = append(valueInfos, ValueInfo{valueLen, value})
		}
		return DataRow{valueInfos}, nil
	case 'I':
		return EmptyQueryResponse{}, nil
	case 'E':
		fallthrough
	case 'N':
		errorInfos := make([]ErrorInfo, 0)
		for {
			code := readInt8(c)
			if code == 0 {
				break
			}
			value := readString(c)
			errorInfos = append(errorInfos, ErrorInfo{code, value})
		}
		if header[0] == 'E' {
			return ErrorResponse{errorInfos}, nil
		} else {
			return NoticeResponse{errorInfos}, nil
		}
	}
	return nil, nil
}

func readAuthenticationMessage(c *PGConn, length uint32) (interface{}, error) {
	subtype := readInt32(c)

	switch subtype {
	case 0:
		return AuthenticationOk{}, nil
	case 2:
		return AuthenticationKerberosV5{}, nil
	case 3:
		return AuthenticationCleartextPassword{}, nil
	case 5:
		var salt [4]byte
		for i := 0; i < 4; i++ {
			salt[i] = readInt8(c)
		}
		return AuthenticationMD5Password{salt}, nil
	case 6:
		return AuthenticationSCMCredential{}, nil
	case 7:
		return AuthenticationGSS{}, nil
	case 9:
		return AuthenticationSSPI{}, nil
	case 8:
		data := readByteN(c, int(length)-8)
		return AuthenticationGSSContinue{data}, nil
	}
	return nil, nil
}

func StringField(fi FieldInfo) string {
	switch fi.DatatypeObjectId {
	case 0x17:
		return "int"
	case 0x412:
		return fmt.Sprintf("char(%d)", fi.TypeModifier-4)
	case 0x413:
		return fmt.Sprintf("varchar(%d)", fi.TypeModifier-4)
	case 0x43a:
		return "date"
	case 0x43f9:
		return "enum"
	}
	return fmt.Sprintf("unknown DatatypeObjectId (%x)", fi.DatatypeObjectId)
}

func StringRows(drs []DataRow, rd RowDescription) string {
	var buffer bytes.Buffer
	for i := 0; i < len(drs); i++ {
		buffer.WriteString(StringRow(drs[i], rd))
		buffer.WriteString("\n")
	}
	if buffer.Len() > 0 {
		buffer.Truncate(buffer.Len() - 1)
	}
	return buffer.String()
}

func StringRow(dr DataRow, rd RowDescription) string {
	var buffer bytes.Buffer
	buffer.WriteString("{ ")
	for i := 0; i < len(dr.ValueInfos); i++ {
		vi := dr.ValueInfos[i]
		fi := rd.FieldInfos[i]
		switch fi.DatatypeObjectId {
		case 0x17:
			buffer.WriteString(fmt.Sprintf("%d, ",
				getIntFromByteSlice(vi.Value, int(vi.ValueLen))))
		case 0x412:
			buffer.WriteString(fmt.Sprintf("\"%s\", ", string(vi.Value)))
		case 0x413:
			buffer.WriteString(fmt.Sprintf("\"%s\", ", string(vi.Value)))
		case 0x43a:
			buffer.WriteString(fmt.Sprintf("\"%s\", ", string(vi.Value)))
		case 0x43f9:
			buffer.WriteString(fmt.Sprintf("'%c', ", vi.Value[0]))
		}
	}
	if buffer.Len() > 2 {
		buffer.Truncate(buffer.Len() - 2)
	} else {
		buffer.Truncate(buffer.Len() - 1)
	}
	buffer.WriteString(" }")
	return buffer.String()
}
