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

export interface confas_confas_next {
  __typename: "ConfaCursor";
  id: string | null;
  createdAt: string | null;
}

export interface confas_confas {
  __typename: "Confas";
  items: confas_confas_items[];
  next: confas_confas_next | null;
}

export interface confas {
  confas: confas_confas;
}

export interface confasVariables {
  where: ConfaLookup;
  limit: number;
  cursor?: ConfaCursorInput | null;
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
  request: ConfaUpdate;
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
  request: ConfaUpdate;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: events
// ====================================================

export interface events_events_items {
  __typename: "Event";
  id: string;
  roomId: string;
  createdAt: string;
  payload: string;
}

export interface events_events_next {
  __typename: "EventCursor";
  id: string | null;
  createdAt: string | null;
}

export interface events_events {
  __typename: "Events";
  items: events_events_items[];
  next: events_events_next | null;
}

export interface events {
  events: events_events;
}

export interface eventsVariables {
  where: EventLookup;
  limit: number;
  cursor?: EventCursorInput | null;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: profiles
// ====================================================

export interface profiles_profiles_items_avatarThumbnail {
  __typename: "Image";
  format: string;
  data: string;
}

export interface profiles_profiles_items {
  __typename: "Profile";
  id: string;
  ownerId: string;
  handle: string;
  displayName: string | null;
  avatarThumbnail: profiles_profiles_items_avatarThumbnail | null;
}

export interface profiles_profiles_next {
  __typename: "ProfileCursor";
  id: string | null;
}

export interface profiles_profiles {
  __typename: "Profiles";
  items: profiles_profiles_items[];
  next: profiles_profiles_next | null;
}

export interface profiles {
  profiles: profiles_profiles;
}

export interface profilesVariables {
  where: ProfileLookup;
  limit: number;
  cursor?: ProfileCursorInput | null;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: updateProfile
// ====================================================

export interface updateProfile_updateProfile {
  __typename: "Profile";
  id: string;
  ownerId: string;
  handle: string;
  displayName: string | null;
}

export interface updateProfile {
  updateProfile: updateProfile_updateProfile;
}

export interface updateProfileVariables {
  request: ProfileUpdate;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: requestAvatarUpload
// ====================================================

export interface requestAvatarUpload_requestAvatarUpload {
  __typename: "UploadToken";
  url: string;
  formData: string;
}

export interface requestAvatarUpload {
  requestAvatarUpload: requestAvatarUpload_requestAvatarUpload;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: recordings
// ====================================================

export interface recordings_recordings_items {
  __typename: "Recording";
  key: string;
  roomId: string;
  status: RecordingStatus;
  createdAt: number;
  startedAt: number;
  stoppedAt: number | null;
}

export interface recordings_recordings_nextFrom {
  __typename: "RecordingFrom";
  key: string;
}

export interface recordings_recordings {
  __typename: "Recordings";
  items: recordings_recordings_items[];
  nextFrom: recordings_recordings_nextFrom | null;
}

export interface recordings {
  recordings: recordings_recordings;
}

export interface recordingsVariables {
  where: RecordingLookup;
  limit: number;
  from?: RecordingFromInput | null;
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
  state: TalkState;
}

export interface talksHydrated_talks_next {
  __typename: "TalkCursor";
  id: string | null;
  createdAt: string | null;
}

export interface talksHydrated_talks {
  __typename: "Talks";
  items: talksHydrated_talks_items[];
  next: talksHydrated_talks_next | null;
}

export interface talksHydrated {
  talks: talksHydrated_talks;
}

export interface talksHydratedVariables {
  where: TalkLookup;
  limit: number;
  cursor?: TalkCursorInput | null;
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
  title: string;
  state: TalkState;
}

export interface talks_talks_next {
  __typename: "TalkCursor";
  id: string | null;
  createdAt: string | null;
}

export interface talks_talks {
  __typename: "Talks";
  items: talks_talks_items[];
  next: talks_talks_next | null;
}

export interface talks {
  talks: talks_talks;
}

export interface talksVariables {
  where: TalkLookup;
  limit: number;
  cursor?: TalkCursorInput | null;
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
  state: TalkState;
}

export interface createTalk {
  createTalk: createTalk_createTalk;
}

export interface createTalkVariables {
  where: ConfaLookup;
  request: TalkUpdate;
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
  state: TalkState;
}

export interface updateTalk {
  updateTalk: updateTalk_updateTalk;
}

export interface updateTalkVariables {
  where: TalkLookup;
  request: TalkUpdate;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: startTalkRecording
// ====================================================

export interface startTalkRecording_startTalkRecording {
  __typename: "Talk";
  id: string;
  ownerId: string;
  confaId: string;
  roomId: string;
  handle: string;
  title: string;
  state: TalkState;
}

export interface startTalkRecording {
  startTalkRecording: startTalkRecording_startTalkRecording;
}

export interface startTalkRecordingVariables {
  where: TalkLookup;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: stopTalkRecording
// ====================================================

export interface stopTalkRecording_stopTalkRecording {
  __typename: "Talk";
  id: string;
  ownerId: string;
  confaId: string;
  roomId: string;
  handle: string;
  title: string;
  state: TalkState;
}

export interface stopTalkRecording {
  stopTalkRecording: stopTalkRecording_stopTalkRecording;
}

export interface stopTalkRecordingVariables {
  where: TalkLookup;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

//==============================================================
// START Enums and Input Objects
//==============================================================

export enum RecordingStatus {
  PROCESSING = "PROCESSING",
  READY = "READY",
  RECORDING = "RECORDING",
}

export enum TalkState {
  CREATED = "CREATED",
  ENDED = "ENDED",
  LIVE = "LIVE",
  RECORDING = "RECORDING",
}

export interface ConfaCursorInput {
  id?: string | null;
  createdAt?: string | null;
  Asc?: boolean | null;
}

export interface ConfaLookup {
  id?: string | null;
  ownerId?: string | null;
  handle?: string | null;
}

export interface ConfaUpdate {
  handle?: string | null;
  title?: string | null;
  description?: string | null;
}

export interface EventCursorInput {
  id?: string | null;
  createdAt?: string | null;
  Asc?: boolean | null;
}

export interface EventLookup {
  roomId: string;
}

export interface ProfileCursorInput {
  id?: string | null;
}

export interface ProfileLookup {
  ownerIds?: string[] | null;
  handle?: string | null;
}

export interface ProfileUpdate {
  handle?: string | null;
  displayName?: string | null;
}

export interface RecordingFromInput {
  key: string;
}

export interface RecordingLookup {
  roomId: string;
  key?: string | null;
}

export interface TalkCursorInput {
  id?: string | null;
  createdAt?: string | null;
  Asc?: boolean | null;
}

export interface TalkLookup {
  id?: string | null;
  ownerId?: string | null;
  speakerId?: string | null;
  confaId?: string | null;
  handle?: string | null;
}

export interface TalkUpdate {
  handle?: string | null;
  title?: string | null;
  description?: string | null;
}

//==============================================================
// END Enums and Input Objects
//==============================================================
