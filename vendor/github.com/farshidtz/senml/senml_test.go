package senml

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"testing"
)

func ExampleEncode1() {
	v := 23.1
	var p Pack = []Record{
		Record{Value: &v, Unit: "Cel", Name: "urn:dev:ow:10e2073a01080063"},
	}

	dataOut, err := p.Encode(JSON, OutputOptions{})
	if err != nil {
		fmt.Println("Encode of SenML failed")
	} else {
		fmt.Println(string(dataOut))
	}
	// Output: [{"n":"urn:dev:ow:10e2073a01080063","u":"Cel","v":23.1}]
}

func ExampleEncode2() {
	v1 := 23.5
	v2 := 23.6
	var p Pack = []Record{
		Record{Value: &v1, Unit: "Cel", BaseName: "urn:dev:ow:10e2073a01080063", Time: 1.276020076305e+09},
		Record{Value: &v2, Unit: "Cel", Time: 1.276020091305e+09},
	}

	dataOut, err := p.Encode(JSON, OutputOptions{})
	if err != nil {
		fmt.Println("Encode of SenML failed")
	} else {
		fmt.Println(string(dataOut))
	}
	// Output: [{"bn":"urn:dev:ow:10e2073a01080063","u":"Cel","t":1276020076.305,"v":23.5},{"u":"Cel","t":1276020091.305,"v":23.6}]
}

type TestVector struct {
	testDecode bool
	format     Format
	binary     bool
	value      string
}

var testVectors = []TestVector{
	{true, JSON, false, "W3siYm4iOiJkZXYxMjMiLCJidCI6LTQ1LjY3LCJidSI6ImRlZ0MiLCJidmVyIjo1LCJuIjoidGVtcCIsInUiOiJkZWdDIiwidCI6LTEsInV0IjoxMCwidiI6MjIuMSwicyI6MH0seyJuIjoicm9vbSIsInQiOi0xLCJ2cyI6ImtpdGNoZW4ifSx7Im4iOiJkYXRhIiwidmQiOiJhYmMifSx7Im4iOiJvayIsInZiIjp0cnVlfV0="},
	{true, CBOR, true, "hKpiYm5mZGV2MTIzYmJ0+8BG1cKPXCj2YmJ1ZGRlZ0NkYnZlcgVhbmR0ZW1wYXP7AAAAAAAAAABhdPu/8AAAAAAAAGF1ZGRlZ0NidXT7QCQAAAAAAABhdvtANhmZmZmZmqNhbmRyb29tYXT7v/AAAAAAAABidnNna2l0Y2hlbqJhbmRkYXRhYnZkY2FiY6JhbmJva2J2YvU="},
	{true, XML, false, "PHNlbnNtbCB4bWxucz0idXJuOmlldGY6cGFyYW1zOnhtbDpuczpzZW5tbCI+PHNlbm1sIGJuPSJkZXYxMjMiIGJ0PSItNDUuNjciIGJ1PSJkZWdDIiBidmVyPSI1IiBuPSJ0ZW1wIiB1PSJkZWdDIiB0PSItMSIgdXQ9IjEwIiB2PSIyMi4xIiBzPSIwIj48L3Nlbm1sPjxzZW5tbCBuPSJyb29tIiB0PSItMSIgdnM9ImtpdGNoZW4iPjwvc2VubWw+PHNlbm1sIG49ImRhdGEiIHZkPSJhYmMiPjwvc2VubWw+PHNlbm1sIG49Im9rIiB2Yj0idHJ1ZSI+PC9zZW5tbD48L3NlbnNtbD4="},
	{false, CSV, false, "dGVtcCwyNTU2OC45OTk5ODgsMjIuMTAwMDAwLGRlZ0MNCg=="},
	{true, MPACK, true, "lIqiYm6mZGV2MTIzomJ0y8BG1cKPXCj2omJ1pGRlZ0OkYnZlcgWhbqR0ZW1woXPLAAAAAAAAAAChdMu/8AAAAAAAAKF1pGRlZ0OidXTLQCQAAAAAAAChdstANhmZmZmZmoOhbqRyb29toXTLv/AAAAAAAACidnOna2l0Y2hlboKhbqRkYXRhonZko2FiY4KhbqJva6J2YsM="},
	{false, LINEP, false, "Zmx1ZmZ5U2VubWwsbj10ZW1wLHU9ZGVnQyB2PTIyLjEgLTEwMDAwMDAwMDAK"},
}

func TestEncode(t *testing.T) {
	value := 22.1
	sum := 0.0
	vb := true
	var pack Pack = []Record{
		Record{BaseName: "dev123",
			BaseTime:    -45.67,
			BaseUnit:    "degC",
			BaseVersion: 5,
			Value:       &value, Unit: "degC", Name: "temp", Time: -1.0, UpdateTime: 10.0, Sum: &sum},
		Record{StringValue: "kitchen", Name: "room", Time: -1.0},
		Record{DataValue: "abc", Name: "data"},
		Record{BoolValue: &vb, Name: "ok"},
	}

	options := OutputOptions{Topic: "fluffySenml", PrettyPrint: false}
	for i, vector := range testVectors {

		dataOut, err := pack.Encode(vector.format, options)
		if err != nil {
			t.Fail()
		}
		if vector.binary {
			fmt.Print("Test Encode " + strconv.Itoa(i) + " got: ")
			fmt.Println(dataOut)
		} else {
			fmt.Println("Test Encode " + strconv.Itoa(i) + " got: " + string(dataOut))
		}

		if base64.StdEncoding.EncodeToString(dataOut) != vector.value {
			t.Errorf("Failed Encode for format %d. Got:\n%s", i, string(dataOut))
			decoded, err := base64.StdEncoding.DecodeString(vector.value)
			if err != nil {
				t.Fatalf("Error decoding test value: %s", err)
			}
			t.Errorf("Expected:\n%s", string(decoded))
		}
	}

}

func TestDecode(t *testing.T) {
	for i, vector := range testVectors {
		t.Logf("Doing TestDecode for vector %d", i)

		if vector.testDecode {
			data, err := base64.StdEncoding.DecodeString(vector.value)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("%s", data)
			pack, err := Decode(data, vector.format)
			if err != nil {
				t.Fatal(err)
			}

			dataOut, err := pack.Encode(JSON, OutputOptions{PrettyPrint: true})
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("Test Decode %d. Got:\n%s\n", i, string(dataOut))
		}
	}
}

func TestNormalize(t *testing.T) {
	value := 22.1
	sum := 0.0
	vb := true
	var pack Pack = []Record{
		Record{BaseName: "dev123/",
			BaseTime:    897845.67,
			BaseUnit:    "degC",
			BaseVersion: 5,
			Value:       &value, Unit: "degC", Name: "temp", Time: -1.0, UpdateTime: 10.0, Sum: &sum},
		Record{StringValue: "kitchen", Name: "room", Time: -1.0},
		Record{DataValue: "abc", Name: "data"},
		Record{BoolValue: &vb, Name: "ok"},
	}

	normalized := pack.Normalize()

	dataOut, err := normalized.Encode(JSON, OutputOptions{PrettyPrint: true})
	if err != nil {
		t.Fail()
	}
	fmt.Println("Test Normalize got: " + string(dataOut))

	testValue := "WwogIHsiYnZlciI6NSwibiI6ImRldjEyMy90ZW1wIiwidSI6ImRlZ0MiLCJ0Ijo4OTc4NDQuNjcsInV0IjoxMCwidiI6MjIuMSwicyI6MH0sCiAgeyJidmVyIjo1LCJuIjoiZGV2MTIzL3Jvb20iLCJ1IjoiZGVnQyIsInQiOjg5Nzg0NC42NywidnMiOiJraXRjaGVuIn0sCiAgeyJidmVyIjo1LCJuIjoiZGV2MTIzL2RhdGEiLCJ1IjoiZGVnQyIsInQiOjg5Nzg0NS42NywidmQiOiJhYmMifSwKICB7ImJ2ZXIiOjUsIm4iOiJkZXYxMjMvb2siLCJ1IjoiZGVnQyIsInQiOjg5Nzg0NS42NywidmIiOnRydWV9Cl0K"
	if base64.StdEncoding.EncodeToString(dataOut) != testValue {
		t.Errorf("Failed Normalize got:\n%v", string(dataOut))
		decoded, err := base64.StdEncoding.DecodeString(testValue)
		if err != nil {
			t.Fatalf("Error decoding test value: %s", err)
		}
		t.Errorf("Expected:\n%v", string(decoded))
	}
}

func TestBadInput1(t *testing.T) {
	data := []byte(" foo ")
	_, err := Decode(data, JSON)
	if err == nil {
		t.Fail()
	}
}

func TestBadInput2(t *testing.T) {
	data := []byte(" { \"n\":\"hi\" } ")
	_, err := Decode(data, JSON)
	if err == nil {
		t.Fail()
	}
}

func TestBadInputNoValue(t *testing.T) {
	data := []byte("  [ { \"n\":\"hi\" } ] ")
	_, err := Decode(data, JSON)
	if err == nil {
		t.Fail()
	}
}

func TestInputNumericName(t *testing.T) {
	data := []byte("  [ { \"n\":\"3a\", \"v\":1.0 } ] ")
	_, err := Decode(data, JSON)
	if err != nil {
		t.Fail()
	}
}

func TestBadInputNumericName(t *testing.T) {
	data := []byte("  [ { \"n\":\"-3b\", \"v\":1.0 } ] ")
	_, err := Decode(data, JSON)
	if err == nil {
		t.Fail()
	}
}

func TestInputWeirdName(t *testing.T) {
	data := []byte("  [ { \"n\":\"Az3-:./_\", \"v\":1.0 } ] ")
	_, err := Decode(data, JSON)
	if err != nil {
		t.Fail()
	}
}

func TestBadInputWeirdName(t *testing.T) {
	data := []byte("  [ { \"n\":\"A;b\", \"v\":1.0 } ] ")
	_, err := Decode(data, JSON)
	if err == nil {
		t.Fail()
	}
}

func TestInputWeirdBaseName(t *testing.T) {
	data := []byte("[ { \"bn\": \"a\" , \"n\":\"/b\" , \"v\":1.0} ] ")
	_, err := Decode(data, JSON)
	if err != nil {
		t.Fail()
	}
}

func TestBadInputNumericBaseName(t *testing.T) {
	data := []byte("[ { \"bn\": \"/3h\" , \"n\":\"i\" , \"v\":1.0} ] ")
	_, err := Decode(data, JSON)
	if err == nil {
		t.Fail()
	}
}

// TODO add
//func TestBadInputUnknownMtuField(t *testing.T) {
//	data := []byte("[ { \"n\":\"hi\", \"v\":1.0, \"mtu_\":1.0  } ] ")
//	_ , err := Decode(data, JSON)
//	if err == nil {
//		t.Fail()
//	}
//}

func TestInputSumOnly(t *testing.T) {
	data := []byte("[ { \"n\":\"a\", \"s\":1.0 } ] ")
	_, err := Decode(data, JSON)
	if err != nil {
		t.Fail()
	}
}

func TestInputBoolean(t *testing.T) {
	data := []byte("[ { \"n\":\"a\", \"vd\": \"aGkgCg\" } ] ")
	_, err := Decode(data, JSON)
	if err != nil {
		t.Fail()
	}
}

func TestInputData(t *testing.T) {
	data := []byte("  [ { \"n\":\"a\", \"vb\": true } ] ")
	_, err := Decode(data, JSON)
	if err != nil {
		t.Fail()
	}
}

func TestInputString(t *testing.T) {
	data := []byte("  [ { \"n\":\"a\", \"vs\": \"Hi\" } ] ")
	_, err := Decode(data, JSON)
	if err != nil {
		t.Fail()
	}
}
