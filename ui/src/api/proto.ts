import {
  create,
  fromJson,
  toJson,
  type DescMessage,
  type JsonValue,
  type MessageInitShape,
  type MessageShape,
} from '@bufbuild/protobuf';

export const decodeProto = <Desc extends DescMessage>(
  schema: Desc,
  value: unknown
): MessageShape<Desc> => fromJson(schema, value as JsonValue);

export const encodeProto = <Desc extends DescMessage>(
  schema: Desc,
  init: MessageInitShape<Desc>
): JsonValue => toJson(schema, create(schema, init));
