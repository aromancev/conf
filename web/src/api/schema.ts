/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: login
// ====================================================

export interface login {
  login: string;
}

export interface loginVariables {
  address: string;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: createSession
// ====================================================

export interface createSession_createSession {
  __typename: "Token";
  token: string;
  expiresIn: number;
}

export interface createSession {
  createSession: createSession_createSession;
}

export interface createSessionVariables {
  emailToken: string;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: token
// ====================================================

export interface token_token {
  __typename: "Token";
  token: string;
  expiresIn: number;
}

export interface token {
  token: token_token;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: confas
// ====================================================

export interface confas_confas_items {
  __typename: "Confa";
  id: string;
  ownerId: string;
  handle: string;
  title: string;
  description: string;
}

export interface confas_confas {
  __typename: "Confas";
  items: confas_confas_items[];
  nextFrom: string;
}

export interface confas {
  confas: confas_confas;
}

export interface confasVariables {
  where: ConfaInput;
  from?: string | null;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: createConfa
// ====================================================

export interface createConfa_createConfa {
  __typename: "Confa";
  id: string;
  ownerId: string;
  handle: string;
  title: string;
  description: string;
}

export interface createConfa {
  createConfa: createConfa_createConfa;
}

export interface createConfaVariables {
  request: ConfaInput;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: updateConfa
// ====================================================

export interface updateConfa {
  updateConfa: number;
}

export interface updateConfaVariables {
  where: ConfaInput;
  request: ConfaInput;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: events
// ====================================================

export interface events_events_items_payload {
  __typename: "EventPayload";
  type: string;
  payload: string;
}

export interface events_events_items {
  __typename: "Event";
  id: string;
  ownerId: string;
  roomId: string;
  createdAt: string;
  payload: events_events_items_payload;
}

export interface events_events_nextFrom {
  __typename: "EventFrom";
  id: string;
  createdAt: string;
}

export interface events_events {
  __typename: "Events";
  items: events_events_items[];
  nextFrom: events_events_nextFrom | null;
}

export interface events {
  events: events_events;
}

export interface eventsVariables {
  where: EventInput;
  from?: EventFromInput | null;
  limit: EventLimit;
  order?: EventOrder | null;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: talksHydrated
// ====================================================

export interface talksHydrated_talks_items {
  __typename: "Talk";
  id: string;
  ownerId: string;
  confaId: string;
  roomId: string;
  handle: string;
}

export interface talksHydrated_talks {
  __typename: "Talks";
  items: talksHydrated_talks_items[];
  nextFrom: string;
}

export interface talksHydrated {
  talks: talksHydrated_talks;
}

export interface talksHydratedVariables {
  where: TalkInput;
  from?: string | null;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: talks
// ====================================================

export interface talks_talks_items {
  __typename: "Talk";
  id: string;
  ownerId: string;
  confaId: string;
  roomId: string;
  handle: string;
}

export interface talks_talks {
  __typename: "Talks";
  items: talks_talks_items[];
  nextFrom: string;
}

export interface talks {
  talks: talks_talks;
}

export interface talksVariables {
  where: TalkInput;
  from?: string | null;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: createTalk
// ====================================================

export interface createTalk_createTalk {
  __typename: "Talk";
  id: string;
  ownerId: string;
  confaId: string;
  roomId: string;
  handle: string;
}

export interface createTalk {
  createTalk: createTalk_createTalk;
}

export interface createTalkVariables {
  confaId: string;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: startTalk
// ====================================================

export interface startTalk {
  startTalk: string;
}

export interface startTalkVariables {
  talkId: string;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

//==============================================================
// START Enums and Input Objects
//==============================================================

export enum EventOrder {
  ASC = "ASC",
  DESC = "DESC",
}

export interface ConfaInput {
  id?: string | null;
  ownerId?: string | null;
  handle?: string | null;
  title?: string | null;
  description?: string | null;
}

export interface EventFromInput {
  id: string;
  createdAt: string;
}

export interface EventInput {
  roomId: string;
}

export interface EventLimit {
  count: number;
  seconds: number;
}

export interface TalkInput {
  id?: string | null;
  ownerId?: string | null;
  speakerId?: string | null;
  confaId?: string | null;
  handle?: string | null;
}

//==============================================================
// END Enums and Input Objects
//==============================================================
