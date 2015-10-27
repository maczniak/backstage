package postgresql

import (
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
)

type PGConn struct {
	conn *net.TCPConn
	buf  []byte
	idx  int
}

func Connect(address string) (*PGConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	return &PGConn{conn, nil, 0}, nil
}

func (c *PGConn) Login(user, password, database string) (map[string]string, error) {
	sm := StartupMessage{user, database}
	bin, _ := sm.MarshalBinary()
	_, err := c.conn.Write(bin)
	if err != nil {
		return nil, err
	}

	smr, _ := ReadMessage(c)
	switch smr.(type) {
	case AuthenticationMD5Password:
		amp := smr.(AuthenticationMD5Password)
		passdata := "md5" + fmt.Sprintf("%x",
			md5.Sum(append([]byte(fmt.Sprintf("%x",
				md5.Sum([]byte(password+user)))), amp.Salt[:]...)))
		pm := PasswordMessage{passdata}
		bin, _ := pm.MarshalBinary()
		_, err = c.conn.Write(bin)
		if err != nil {
			return nil, err
		}

		ao, _ := ReadMessage(c)
		if _, ok := ao.(AuthenticationOk); !ok {
			log.Fatal(fmt.Errorf("no AuthenticationOk (wrong login info)"))
		}
	default:
		log.Fatal(fmt.Errorf("unknown smr type (%T)", smr))
	}

	ret := make(map[string]string)

authLoop:
	for {
		authResp, _ := ReadMessage(c)
		switch authResp.(type) {
		case ParameterStatus:
			ps := authResp.(ParameterStatus)
			ret[ps.Name] = ps.Value
		case BackendKeyData:
			bkd := authResp.(BackendKeyData)
			ret["_ProcessId"] = strconv.Itoa(int(bkd.ProcessId))
			ret["_SecretKey"] = strconv.Itoa(int(bkd.SecretKey))
		case ReadyForQuery:
			rfq := authResp.(ReadyForQuery)
			ret["_TransactionStatus"] = string(rfq.TransactionStatus)
			break authLoop
		case ErrorResponse:
			return nil, errors.New(authResp.(ErrorResponse).String())
		default:
			return nil, fmt.Errorf("unknown authResp type (%T)", authResp)
		}
	}
	return ret, nil
}

func (c *PGConn) Query(query string) (map[string]interface{}, error) {
	q := Query{query}
	bin, _ := q.MarshalBinary()
	_, err := c.conn.Write(bin)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	var rowDesc RowDescription
	dataRows := make([]DataRow, 0)

queryLoop:
	for {
		queryResp, _ := ReadMessage(c)
		switch queryResp.(type) {
		case RowDescription:
			rowDesc = queryResp.(RowDescription)
			result["description"] = rowDesc
		case DataRow:
			dataRows = append(dataRows, queryResp.(DataRow))
		case CommandComplete:
			cc := queryResp.(CommandComplete)
			result["command_tag"] = cc.CommandTag
		case ReadyForQuery:
			rfq := queryResp.(ReadyForQuery)
			result["transaction_status"] = string(rfq.TransactionStatus)
			break queryLoop
		case ErrorResponse:
			return result, errors.New(queryResp.(ErrorResponse).String())
		default:
			return result, fmt.Errorf("unknown queryResp type (%T)", queryResp)
		}
	}
	result["rows"] = dataRows
	return result, nil
}
