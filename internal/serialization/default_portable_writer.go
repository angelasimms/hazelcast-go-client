package serialization

import (
	. "github.com/hazelcast/go-client/internal/common"
	. "github.com/hazelcast/go-client/internal/serialization/api"
)

type DefaultPortableWriter struct {
	serializer      *PortableSerializer
	output          PositionalDataOutput
	classDefinition *ClassDefinition
	begin           int32
	offset          int32
}

func NewDefaultPortableWriter(serializer *PortableSerializer, output PositionalDataOutput, classDefinition *ClassDefinition) *DefaultPortableWriter {
	begin := output.Position()
	output.WriteZeroBytes(4)
	output.WriteInt32(int32(len(classDefinition.fields)))
	offset := output.Position()
	fieldIndexesLength := (len(classDefinition.fields) + 1) * INT_SIZE_IN_BYTES
	output.WriteZeroBytes(fieldIndexesLength)
	return &DefaultPortableWriter{serializer, output, classDefinition, begin, offset}
}

func (pw *DefaultPortableWriter) WriteByte(fieldName string, value byte) {
	pw.setPosition(fieldName, BYTE)
	pw.output.WriteByte(value)
}

func (pw *DefaultPortableWriter) WriteBool(fieldName string, value bool) {
	pw.setPosition(fieldName, BOOLEAN)
	pw.output.WriteBool(value)
}

func (pw *DefaultPortableWriter) WriteUInt16(fieldName string, value uint16) {
	pw.setPosition(fieldName, CHAR)
	pw.output.WriteUInt16(value)
}

func (pw *DefaultPortableWriter) WriteInt16(fieldName string, value int16) {
	pw.setPosition(fieldName, SHORT)
	pw.output.WriteInt16(value)
}

func (pw *DefaultPortableWriter) WriteInt32(fieldName string, value int32) {
	pw.setPosition(fieldName, INT)
	pw.output.WriteInt32(value)
}

func (pw *DefaultPortableWriter) WriteInt64(fieldName string, value int64) {
	pw.setPosition(fieldName, LONG)
	pw.output.WriteInt64(value)
}

func (pw *DefaultPortableWriter) WriteFloat32(fieldName string, value float32) {
	pw.setPosition(fieldName, FLOAT)
	pw.output.WriteFloat32(value)
}

func (pw *DefaultPortableWriter) WriteFloat64(fieldName string, value float64) {
	pw.setPosition(fieldName, DOUBLE)
	pw.output.WriteFloat64(value)
}

func (pw *DefaultPortableWriter) WriteUTF(fieldName string, value string) {
	pw.setPosition(fieldName, UTF)
	pw.output.WriteUTF(value)
}

func (pw *DefaultPortableWriter) WritePortable(fieldName string, portable Portable) error {
	fieldDefinition := pw.setPosition(fieldName, PORTABLE)
	isNullPortable := portable == nil
	pw.output.WriteBool(isNullPortable)
	pw.output.WriteInt32(fieldDefinition.factoryId)
	pw.output.WriteInt32(fieldDefinition.classId)
	if !isNullPortable {
		pw.serializer.WriteObject(pw.output, portable)
	}
	return nil
}

func (pw *DefaultPortableWriter) WriteNilPortable(fieldName string, factoryId int32, classId int32) error {
	pw.setPosition(fieldName, PORTABLE)
	pw.output.WriteBool(true)
	pw.output.WriteInt32(factoryId)
	pw.output.WriteInt32(classId)
	return nil
}

func (pw *DefaultPortableWriter) WriteByteArray(fieldName string, array []byte) {
	pw.setPosition(fieldName, BYTE_ARRAY)
	pw.output.WriteByteArray(array)
}

func (pw *DefaultPortableWriter) WriteBoolArray(fieldName string, array []bool) {
	pw.setPosition(fieldName, BOOLEAN_ARRAY)
	pw.output.WriteBoolArray(array)
}

func (pw *DefaultPortableWriter) WriteUInt16Array(fieldName string, array []uint16) {
	pw.setPosition(fieldName, CHAR_ARRAY)
	pw.output.WriteUInt16Array(array)
}

func (pw *DefaultPortableWriter) WriteInt16Array(fieldName string, array []int16) {
	pw.setPosition(fieldName, SHORT_ARRAY)
	pw.output.WriteInt16Array(array)
}

func (pw *DefaultPortableWriter) WriteInt32Array(fieldName string, array []int32) {
	pw.setPosition(fieldName, INT_ARRAY)
	pw.output.WriteInt32Array(array)
}

func (pw *DefaultPortableWriter) WriteInt64Array(fieldName string, array []int64) {
	pw.setPosition(fieldName, LONG_ARRAY)
	pw.output.WriteInt64Array(array)
}

func (pw *DefaultPortableWriter) WriteFloat32Array(fieldName string, array []float32) {
	pw.setPosition(fieldName, FLOAT_ARRAY)
	pw.output.WriteFloat32Array(array)
}

func (pw *DefaultPortableWriter) WriteFloat64Array(fieldName string, array []float64) {
	pw.setPosition(fieldName, DOUBLE_ARRAY)
	pw.output.WriteFloat64Array(array)
}

func (pw *DefaultPortableWriter) WriteUTFArray(fieldName string, array []string) {
	pw.setPosition(fieldName, UTF_ARRAY)
	pw.output.WriteUTFArray(array)
}

func (pw *DefaultPortableWriter) WritePortableArray(fieldName string, portableArray []Portable) error {
	var innerOffset int32
	var sample Portable
	fieldDefinition := pw.setPosition(fieldName, PORTABLE_ARRAY)
	var length int32
	if portableArray != nil {
		length = int32(len(portableArray))
	} else {
		length = NULL_ARRAY_LENGTH
	}
	pw.output.WriteInt32(length)
	pw.output.WriteInt32(fieldDefinition.factoryId)
	pw.output.WriteInt32(fieldDefinition.classId)

	if length > 0 {
		innerOffset = pw.output.Position()
		pw.output.WriteZeroBytes(int(length) * 4)
		for i := int32(0); i < length; i++ {
			sample = portableArray[i]
			posVal := pw.output.Position()
			pw.output.PWriteInt32(innerOffset+i*INT_SIZE_IN_BYTES, posVal)
			pw.serializer.WriteObject(pw.output, sample)
		}
	}
	return nil
}

func (pw *DefaultPortableWriter) setPosition(fieldName string, fieldType int32) *FieldDefinition {
	field := pw.classDefinition.fields[fieldName]
	pos := pw.output.Position()
	index := field.index
	pw.output.PWriteInt32(pw.offset+index*INT_SIZE_IN_BYTES, pos)
	pw.output.WriteInt16(int16(len(fieldName)))
	pw.output.WriteBytes(fieldName)
	pw.output.WriteByte(byte(fieldType))
	return field
}

func (pw *DefaultPortableWriter) End() {
	position := pw.output.Position()
	pw.output.PWriteInt32(pw.begin, position)
}
