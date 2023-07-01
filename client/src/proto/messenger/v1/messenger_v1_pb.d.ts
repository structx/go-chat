import * as jspb from 'google-protobuf'



export class Conversation extends jspb.Message {
  getToken(): string;
  setToken(value: string): Conversation;

  getConversation(): string;
  setConversation(value: string): Conversation;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Conversation.AsObject;
  static toObject(includeInstance: boolean, msg: Conversation): Conversation.AsObject;
  static serializeBinaryToWriter(message: Conversation, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Conversation;
  static deserializeBinaryFromReader(message: Conversation, reader: jspb.BinaryReader): Conversation;
}

export namespace Conversation {
  export type AsObject = {
    token: string,
    conversation: string,
  }
}

export class NewEnvelope extends jspb.Message {
  getSender(): string;
  setSender(value: string): NewEnvelope;

  getConversation(): string;
  setConversation(value: string): NewEnvelope;

  getMessage(): string;
  setMessage(value: string): NewEnvelope;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): NewEnvelope.AsObject;
  static toObject(includeInstance: boolean, msg: NewEnvelope): NewEnvelope.AsObject;
  static serializeBinaryToWriter(message: NewEnvelope, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): NewEnvelope;
  static deserializeBinaryFromReader(message: NewEnvelope, reader: jspb.BinaryReader): NewEnvelope;
}

export namespace NewEnvelope {
  export type AsObject = {
    sender: string,
    conversation: string,
    message: string,
  }
}

export class Envelope extends jspb.Message {
  getUid(): string;
  setUid(value: string): Envelope;

  getSender(): string;
  setSender(value: string): Envelope;

  getMessage(): string;
  setMessage(value: string): Envelope;

  getStatus(): SEND_ENVELOPE_STATUS;
  setStatus(value: SEND_ENVELOPE_STATUS): Envelope;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Envelope.AsObject;
  static toObject(includeInstance: boolean, msg: Envelope): Envelope.AsObject;
  static serializeBinaryToWriter(message: Envelope, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Envelope;
  static deserializeBinaryFromReader(message: Envelope, reader: jspb.BinaryReader): Envelope;
}

export namespace Envelope {
  export type AsObject = {
    uid: string,
    sender: string,
    message: string,
    status: SEND_ENVELOPE_STATUS,
  }
}

export enum SEND_ENVELOPE_STATUS { 
  ERROR = 0,
  DELIVERED = 1,
}
