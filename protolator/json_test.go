/*
Copyright IBM Corp. 2017 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

                 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package protolator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/hyperledger/fabric-config/protolator/testprotos"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
)

type testProtoPlainFieldFactory struct {
	fromPrefix string
	toPrefix   string
	fromError  error
	toError    error
}

func (tpff *testProtoPlainFieldFactory) Handles(msg proto.Message, fieldName string, fieldType reflect.Type, fieldValue reflect.Value) bool {
	return fieldName == "plain_field"
}

func (tpff *testProtoPlainFieldFactory) NewProtoField(msg proto.Message, fieldName string, fieldType reflect.Type, fieldValue reflect.Value) (protoField, error) {
	return &plainField{
		baseField: baseField{
			msg:   msg,
			name:  fieldName,
			fType: reflect.TypeOf(""),
			vType: fieldType,
			value: fieldValue,
		},
		populateFrom: func(source interface{}, destType reflect.Type) (reflect.Value, error) {
			sourceAsString := source.(string)
			return reflect.ValueOf(tpff.fromPrefix + sourceAsString), tpff.fromError
		},
		populateTo: func(source reflect.Value) (interface{}, error) {
			return tpff.toPrefix + source.Interface().(string), tpff.toError
		},
	}, nil
}

func TestSimpleMsgPlainField(t *testing.T) {
	gt := NewGomegaWithT(t)

	fromPrefix := "from"
	toPrefix := "to"
	tppff := &testProtoPlainFieldFactory{
		fromPrefix: fromPrefix,
		toPrefix:   toPrefix,
	}

	fieldFactories = []protoFieldFactory{tppff}

	pfValue := "foo"
	startMsg := &testprotos.SimpleMsg{
		PlainField: pfValue,
		MapField:   map[string]string{"1": "2"},
		SliceField: []string{"a", "b"},
	}

	var buffer bytes.Buffer
	err := DeepMarshalJSON(&buffer, startMsg)
	gt.Expect(err).NotTo(HaveOccurred())

	newMsg := &testprotos.SimpleMsg{}
	err = DeepUnmarshalJSON(bytes.NewReader(buffer.Bytes()), newMsg)
	gt.Expect(err).NotTo(HaveOccurred())

	gt.Expect(newMsg.MapField).To(Equal(startMsg.MapField))
	gt.Expect(newMsg.SliceField).To(Equal(startMsg.SliceField))
	gt.Expect(newMsg.PlainField).To(Equal(fromPrefix + toPrefix + startMsg.PlainField))

	tppff.fromError = fmt.Errorf("Failing from intentionally")
	err = DeepUnmarshalJSON(bytes.NewReader(buffer.Bytes()), newMsg)
	gt.Expect(err).To(MatchError("*testprotos.SimpleMsg: error in PopulateFrom for field plain_field for message *testprotos.SimpleMsg: Failing from intentionally"))

	tppff.toError = fmt.Errorf("Failing to intentionally")
	err = DeepMarshalJSON(&buffer, startMsg)
	gt.Expect(err).To(MatchError("*testprotos.SimpleMsg: error in PopulateTo for field plain_field for message *testprotos.SimpleMsg: Failing to intentionally"))
}

type testProtoMapFieldFactory struct {
	fromPrefix string
	toPrefix   string
	fromError  error
	toError    error
}

func (tpff *testProtoMapFieldFactory) Handles(msg proto.Message, fieldName string, fieldType reflect.Type, fieldValue reflect.Value) bool {
	return fieldName == "map_field"
}

func (tpff *testProtoMapFieldFactory) NewProtoField(msg proto.Message, fieldName string, fieldType reflect.Type, fieldValue reflect.Value) (protoField, error) {
	return &mapField{
		baseField: baseField{
			msg:   msg,
			name:  fieldName,
			fType: reflect.TypeOf(""),
			vType: fieldType,
			value: fieldValue,
		},
		populateFrom: func(key string, source interface{}, destType reflect.Type) (reflect.Value, error) {
			sourceAsString := source.(string)
			return reflect.ValueOf(tpff.fromPrefix + key + sourceAsString), tpff.fromError
		},
		populateTo: func(key string, source reflect.Value) (interface{}, error) {
			return tpff.toPrefix + key + source.Interface().(string), tpff.toError
		},
	}, nil
}

func TestSimpleMsgMapField(t *testing.T) {
	gt := NewGomegaWithT(t)

	fromPrefix := "from"
	toPrefix := "to"
	tpmff := &testProtoMapFieldFactory{
		fromPrefix: fromPrefix,
		toPrefix:   toPrefix,
	}
	fieldFactories = []protoFieldFactory{tpmff}

	key := "foo"
	value := "bar"
	startMsg := &testprotos.SimpleMsg{
		PlainField: "1",
		MapField:   map[string]string{key: value},
		SliceField: []string{"a", "b"},
	}

	var buffer bytes.Buffer
	err := DeepMarshalJSON(&buffer, startMsg)
	gt.Expect(err).NotTo(HaveOccurred())

	newMsg := &testprotos.SimpleMsg{}
	err = DeepUnmarshalJSON(bytes.NewReader(buffer.Bytes()), newMsg)
	gt.Expect(err).NotTo(HaveOccurred())

	gt.Expect(newMsg.PlainField).To(Equal(startMsg.PlainField))
	gt.Expect(newMsg.SliceField).To(Equal(startMsg.SliceField))
	gt.Expect(newMsg.MapField[key]).To(Equal(fromPrefix + key + toPrefix + key + startMsg.MapField[key]))

	tpmff.fromError = fmt.Errorf("Failing from intentionally")
	err = DeepUnmarshalJSON(bytes.NewReader(buffer.Bytes()), newMsg)
	gt.Expect(err).To(MatchError("*testprotos.SimpleMsg: error in PopulateFrom for map field map_field with key foo for message *testprotos.SimpleMsg: Failing from intentionally"))

	tpmff.toError = fmt.Errorf("Failing to intentionally")
	err = DeepMarshalJSON(&buffer, startMsg)
	gt.Expect(err).To(MatchError("*testprotos.SimpleMsg: error in PopulateTo for map field map_field and key foo for message *testprotos.SimpleMsg: Failing to intentionally"))
}

type testProtoSliceFieldFactory struct {
	fromPrefix string
	toPrefix   string
	fromError  error
	toError    error
}

func (tpff *testProtoSliceFieldFactory) Handles(msg proto.Message, fieldName string, fieldType reflect.Type, fieldValue reflect.Value) bool {
	return fieldName == "slice_field"
}

func (tpff *testProtoSliceFieldFactory) NewProtoField(msg proto.Message, fieldName string, fieldType reflect.Type, fieldValue reflect.Value) (protoField, error) {
	return &sliceField{
		baseField: baseField{
			msg:   msg,
			name:  fieldName,
			fType: reflect.TypeOf(""),
			vType: fieldType,
			value: fieldValue,
		},
		populateFrom: func(index int, source interface{}, destType reflect.Type) (reflect.Value, error) {
			sourceAsString := source.(string)
			return reflect.ValueOf(tpff.fromPrefix + fmt.Sprintf("%d", index) + sourceAsString), tpff.fromError
		},
		populateTo: func(index int, source reflect.Value) (interface{}, error) {
			return tpff.toPrefix + fmt.Sprintf("%d", index) + source.Interface().(string), tpff.toError
		},
	}, nil
}

func TestSimpleMsgSliceField(t *testing.T) {
	gt := NewGomegaWithT(t)

	fromPrefix := "from"
	toPrefix := "to"
	tpsff := &testProtoSliceFieldFactory{
		fromPrefix: fromPrefix,
		toPrefix:   toPrefix,
	}
	fieldFactories = []protoFieldFactory{tpsff}

	value := "foo"
	startMsg := &testprotos.SimpleMsg{
		PlainField: "1",
		MapField:   map[string]string{"a": "b"},
		SliceField: []string{value},
	}

	var buffer bytes.Buffer
	err := DeepMarshalJSON(&buffer, startMsg)
	gt.Expect(err).NotTo(HaveOccurred())

	newMsg := &testprotos.SimpleMsg{}
	err = DeepUnmarshalJSON(bytes.NewReader(buffer.Bytes()), newMsg)
	gt.Expect(err).NotTo(HaveOccurred())

	gt.Expect(newMsg.PlainField).To(Equal(startMsg.PlainField))
	gt.Expect(newMsg.MapField).To(Equal(startMsg.MapField))
	gt.Expect(newMsg.SliceField[0]).To(Equal(fromPrefix + "0" + toPrefix + "0" + startMsg.SliceField[0]))

	tpsff.fromError = fmt.Errorf("Failing from intentionally")
	err = DeepUnmarshalJSON(bytes.NewReader(buffer.Bytes()), newMsg)
	gt.Expect(err).To(MatchError("*testprotos.SimpleMsg: error in PopulateFrom for slice field slice_field at index 0 for message *testprotos.SimpleMsg: Failing from intentionally"))

	tpsff.toError = fmt.Errorf("Failing to intentionally")
	err = DeepMarshalJSON(&buffer, startMsg)
	gt.Expect(err).To(MatchError("*testprotos.SimpleMsg: error in PopulateTo for slice field slice_field at index 0 for message *testprotos.SimpleMsg: Failing to intentionally"))
}

type testProtoFailFactory struct{}

func (tpff testProtoFailFactory) Handles(msg proto.Message, fieldName string, fieldType reflect.Type, fieldValue reflect.Value) bool {
	return true
}

func (tpff testProtoFailFactory) NewProtoField(msg proto.Message, fieldName string, fieldType reflect.Type, fieldValue reflect.Value) (protoField, error) {
	return nil, fmt.Errorf("Intentionally failing")
}

func TestFailFactory(t *testing.T) {
	gt := NewGomegaWithT(t)

	fieldFactories = []protoFieldFactory{&testProtoFailFactory{}}

	var buffer bytes.Buffer
	err := DeepMarshalJSON(&buffer, &testprotos.SimpleMsg{})
	gt.Expect(err).To(MatchError("*testprotos.SimpleMsg: Intentionally failing"))
}

func TestJSONUnmarshalMaxUint32(t *testing.T) {
	gt := NewGomegaWithT(t)

	fieldName := "numField"
	jsonString := fmt.Sprintf("{\"%s\":%d}", fieldName, math.MaxUint32)
	m, err := jsonToMap([]byte(jsonString))
	gt.Expect(err).NotTo(HaveOccurred())
	gt.Expect(m[fieldName]).To(BeAssignableToTypeOf(json.Number("")))
}

func TestMostlyDeterministicMarshal(t *testing.T) {
	gt := NewGomegaWithT(t)

	multiKeyMap := &testprotos.SimpleMsg{
		MapField: map[string]string{
			"a": "b",
			"c": "d",
			"e": "f",
			"g": "h",
			"i": "j",
			"k": "l",
			"m": "n",
			"o": "p",
			"q": "r",
			"s": "t",
			"u": "v",
			"w": "x",
			"y": "z",
		},
	}

	result, err := MostlyDeterministicMarshal(multiKeyMap)
	gt.Expect(err).NotTo(HaveOccurred())
	gt.Expect(result).NotTo(BeNil())

	// Golang map marshaling is non-deterministic by default, by marshaling
	// the same message with an embedded map multiple times, we should
	// detect a mismatch if the default behavior persists.  Even with 3 map
	// elements, there is usually a mismatch within 2-3 iterations, so 13
	// entries and 10 iterations seems like a reasonable check.
	for i := 0; i < 10; i++ {
		newResult, err := MostlyDeterministicMarshal(multiKeyMap)
		gt.Expect(err).NotTo(HaveOccurred())
		gt.Expect(newResult).To(Equal(result))
	}

	unmarshaled := &testprotos.SimpleMsg{}
	err = proto.Unmarshal(result, unmarshaled)
	gt.Expect(err).NotTo(HaveOccurred())
	gt.Expect(proto.Equal(unmarshaled, multiKeyMap)).To(BeTrue())
}
