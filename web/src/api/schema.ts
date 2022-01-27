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
  where: ConfaLookup;
  limit: number;
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
  request: ConfaMask;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: updateConfa
// ====================================================

export interface updateConfa_updateConfa {
  __typename: "Confa";
  id: string;
  ownerId: string;
  handle: string;
  title: string;
  description: string;
}

export interface updateConfa {
  updateConfa: updateConfa_updateConfa;
}

export interface updateConfaVariables {
  where: ConfaLookup;
  request: ConfaMask;
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
  where: EventLookup;
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
  title: string;
  description: string;
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
  where: TalkLookup;
  limit: number;
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
  where: TalkLookup;
  limit: number;
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
  title: string;
  description: string;
}

export interface createTalk {
  createTalk: createTalk_createTalk;
}

export interface createTalkVariables {
  where: ConfaLookup;
  request: TalkMask;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: updateTalk
// ====================================================

export interface updateTalk_updateTalk {
  __typename: "Talk";
  id: string;
  confaId: string;
  roomId: string;
  ownerId: string;
  handle: string;
  title: string;
  description: string;
}

export interface updateTalk {
  updateTalk: updateTalk_updateTalk;
}

export interface updateTalkVariables {
  where: TalkLookup;
  request: TalkMask;
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

export interface ConfaLookup {
  id?: string | null;
  ownerId?: string | null;
  handle?: string | null;
}

export interface ConfaMask {
  handle?: string | null;
  title?: string | null;
  description?: string | null;
}

export interface EventFromInput {
  id: string;
  createdAt: string;
}

export interface EventLimit {
  count: number;
  seconds: number;
}

export interface EventLookup {
  roomId: string;
}

export interface TalkLookup {
  id?: string | null;
  ownerId?: string | null;
  speakerId?: string | null;
  confaId?: string | null;
  handle?: string | null;
}

export interface TalkMask {
  handle?: string | null;
  title?: string | null;
  description?: string | null;
}

//==============================================================
// END Enums and Input Objects
//==============================================================
