/*Copyright (C) 2017 Alex Beltran

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to:
The Free Software Foundation, Inc.
59 Temple Place - Suite 330
Boston, MA  02111-1307, USA.

As a special exception, if other files instantiate templates or
use macros or inline functions from this file, or you compile
this file and link it with other works to produce a work based
on this file, this file does not by itself cause the resulting
work to be covered by the GNU General Public License. However
the source code for this file must still be made available in
accordance with section (3) of the GNU General Public License.

This exception does not invalidate any other reasons why a work
based on this file might be covered by the GNU General Public
License.
*/
package encoding

import (
	"encoding/json"
	"log"
	"testing"

	bactype "github.com/alexbeltran/gobacnet/types"
)

func TestReadPropertyService(t *testing.T) {
	// This value is based on a known sample
	expected := []byte{129, 10, 0, 22, 1, 36, 9, 124, 1, 29, 255, 0, 5, 1, 12,
		12, 0, 0, 0, 1, 25, 85}

	e := NewEncoder()
	//s := `{"ID":24289,"MaxAPDU":480,"Address":{"Mac":"ChQAzLrA","MacLen":6,"Net":2428,"Adr":"HQ==","AdrLen":1}}`
	var mac []uint8
	var adr []uint8
	json.Unmarshal([]byte("\"ChQAzLrA\""), &mac)
	json.Unmarshal([]byte("\"HQ==\""), &adr)
	readProp := bactype.ReadPropertyData{
		Object: bactype.Object{
			ID: bactype.ObjectID{
				Type:     0,
				Instance: 1,
			},
			Properties: []bactype.Property{
				bactype.Property{
					Type:       85,
					ArrayIndex: 0xFFFFFFFF,
				},
			},
		},
	}

	dest := bactype.Address{
		Net:    2428,
		Mac:    mac,
		MacLen: 6,
		Len:    1,
		Adr:    adr,
	}
	e.NPDU(bactype.NPDU{
		Version:               bactype.ProtocolVersion,
		IsNetworkLayerMessage: false,
		ExpectingReply:        true,
		HopCount:              bactype.DefaultHopCount,
		Priority:              bactype.Normal,
		Destination:           &dest,
	})
	e.ReadProperty(1, readProp)
	data := e.Bytes()

	enc := NewEncoder()
	bv := bactype.BVLC{
		Type:     bactype.BVLCTypeBacnetIP,
		Function: bactype.BacFuncUnicast,
		Length:   4 + uint16(len(data)),
		Data:     data,
	}
	enc.BVLC(bv)

	raw := enc.Bytes()
	for i, b := range raw {
		if expected[i] != b {
			t.Errorf("Error during decoding: %x does not equal expected %x", b, expected[i])
		}
	}
	if len(raw) != len(expected) {
		t.Fatalf("There is a mismatch in sizes. Got: %d, Expected:%d", len(raw), len(expected))
	}
	t.Logf("Length: %d", len(raw))
}

func TestReadPropertyResponse(t *testing.T) {
	// This value is based on a known sample
	in := []byte{48, 1, 12, 12, 0, 0, 0, 1, 25, 85, 62, 68, 192, 160, 0, 0, 63}
	d := NewDecoder(in)
	apdu := bactype.APDU{}
	d.APDU(&apdu)

	rpd := bactype.ReadPropertyData{}
	err := d.ReadProperty(&rpd)
	if err != nil {
		t.Fatal(err)
	}
	apd := NewDecoder(rpd.Object.Properties[0].Data)
	x, err := apd.AppData()
	if err != nil {
		t.Fatal(err)
	}
	f := x.(float32)
	if f != -5.0 {
		t.Fatalf("Final value was not decrypted properly")
	}

}

func TestWhoIs(t *testing.T) {
	e := NewEncoder()
	var low int32 = 28
	var high int32 = 32
	err := e.WhoIs(low, high)
	if err != nil {
		t.Fatal(err)
	}

	d := NewDecoder(e.Bytes())
	a := bactype.APDU{}
	d.APDU(&a)

	d = NewDecoder(a.RawData)
	var lowOut, highOut int32
	d.WhoIs(&lowOut, &highOut)

	if err = d.Error(); err != nil {
		t.Fatal(err)
	}

	if low != lowOut || high != highOut {
		t.Fatalf("WhoIs was not decoded properly. Low was %d, given %d. High was %d, given %d", low, lowOut, high, highOut)
	}
}

func TestIAm(t *testing.T) {
	b := []byte{196, 2, 3, 180, 113, 34, 1, 224, 145, 3, 33, 24}
	dec := NewDecoder(b)
	for dec.len() > 0 {
		x, err := dec.AppData()
		if err != nil {
			t.Fatal(err)
		}

		log.Printf("app: %v", x)
	}
}
