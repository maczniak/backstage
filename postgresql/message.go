// TODO: return result
// TODO: readable output
package postgresql

import (
	"bytes"
	"fmt"
	"strings"
)

//type Message interface {}

/* login responses */

type AuthenticationOk struct{}
type AuthenticationKerberosV5 struct{}
type AuthenticationCleartextPassword struct{}
type AuthenticationMD5Password struct {
	Salt [4]byte
}
type AuthenticationSCMCredential struct{}
type AuthenticationGSS struct{}
type AuthenticationSSPI struct{}
type AuthenticationGSSContinue struct {
	Data []byte
}

type ParameterStatus struct {
	Name  string
	Value string
}

func (p ParameterStatus) String() string {
	return fmt.Sprintf("ParameterStatus(%s:%s)", p.Name, p.Value)
}

type BackendKeyData struct {
	ProcessId uint32
	SecretKey uint32
}

func (b BackendKeyData) String() string {
	return fmt.Sprintf("BackendData(pid: %d, secret: %d)",
		b.ProcessId, b.SecretKey)
}

type ReadyForQuery struct {
	TransactionStatus byte
}

func (r ReadyForQuery) String() string {
	return fmt.Sprintf("ReadyForQuery(%c)", r.TransactionStatus)
}

/* simple query responses */

type CommandComplete struct {
	CommandTag string
}

func (c CommandComplete) String() string {
	return fmt.Sprintf("CommandComplete(\"%s\")", c.CommandTag)
}

type CopyInResponse struct {
	Format      uint8
	ColumnCount uint16
	FormatCodes []uint16
}

type CopyOutResponse struct {
	Format      uint8
	ColumnCount uint16
	FormatCodes []uint16
}

type FieldInfo struct {
	Name               string
	TableObjectId      uint32
	ColumnAttributeNum uint16
	DatatypeObjectId   uint32
	DatatypeSize       uint16
	TypeModifier       uint32
	FormatCode         uint16
}
type RowDescription struct {
	FieldInfos []FieldInfo
}

func (r RowDescription) String() string {
	ret := ""
	for i := 0; i < len(r.FieldInfos); i++ {
		fi := r.FieldInfos[i]
		ret += fmt.Sprintf("%d: %s %s\n", i, fi.Name, StringField(fi))
	}
	return strings.TrimRight(ret, "\n")
}

type ValueInfo struct {
	ValueLen uint32
	Value    []byte
}
type DataRow struct {
	ValueInfos []ValueInfo
}

type EmptyQueryResponse struct{}

type ErrorInfo struct {
	Code  byte
	Value string
}
type ErrorResponse struct {
	ErrorInfos []ErrorInfo
}
type NoticeResponse struct {
	ErrorInfos []ErrorInfo
}

func (e ErrorInfo) String() string {
	var ret string
	switch e.Code {
	case 'S':
		ret = "Severity: "
	case 'C':
		ret = "Code: "
	case 'M':
		ret = "Message: "
	case 'D':
		ret = "Detail: "
	case 'H':
		ret = "Hint: "
	case 'P':
		ret = "Position: "
	case 'p':
		ret = "Internal position: "
	case 'q':
		ret = "Internal query: "
	case 'W':
		ret = "Where: "
	case 's':
		ret = "Schema name: "
	case 't':
		ret = "Table name: "
	case 'c':
		ret = "Column name: "
	case 'd':
		ret = "Data type name: "
	case 'n':
		ret = "Constraint name: "
	case 'F':
		ret = "File: "
	case 'L':
		ret = "Line: "
	case 'R':
		ret = "Routine: "
	}
	return ret + e.Value
}
func (e ErrorResponse) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("Error Info:")
	for i := 0; i < len(e.ErrorInfos); i++ {
		buffer.WriteString("\n\t")
		buffer.WriteString(e.ErrorInfos[i].String())
	}
	return buffer.String()
}

func setLength(data []byte, length int) {
	data[3] = byte(length % 256)
	length >>= 8
	data[2] = byte(length % 256)
	length >>= 8
	data[1] = byte(length % 256)
	length >>= 8
	data[0] = byte(length % 256)
}

/* login requests */

type StartupMessage struct {
	User     string
	Database string
}

func (m *StartupMessage) MarshalBinary() (data []byte, err error) {
	err = nil
	data = []byte{0, 0, 0, 0} // dummy length
	data = append(data, 0, 3, 0, 0)
	data = append(data, []byte("user")...)
	data = append(data, 0)
	data = append(data, []byte(m.User)...)
	data = append(data, 0)
	data = append(data, []byte("database")...)
	data = append(data, 0)
	data = append(data, []byte(m.Database)...)
	data = append(data, 0)
	data = append(data, 0) // end of message
	setLength(data, len(data))
	return
}

type PasswordMessage struct {
	Password string
}

func (m *PasswordMessage) MarshalBinary() (data []byte, err error) {
	err = nil
	data = []byte{'p', 0, 0, 0, 0} // type and dummy length
	data = append(data, []byte(m.Password)...)
	data = append(data, 0)
	setLength(data[1:], len(data)-1)
	return
}

/* simple query requests */

type Query struct {
	Query string
}

func (m *Query) MarshalBinary() (data []byte, err error) {
	err = nil
	data = []byte{'Q', 0, 0, 0, 0} // type and dummy length
	data = append(data, []byte(m.Query)...)
	data = append(data, 0)
	setLength(data[1:], len(data)-1)
	return
}
