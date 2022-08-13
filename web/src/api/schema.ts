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

export interface events_events_items_payload_peerState_tracks {
  __typename: "Track";
  id: string;
  hint: Hint;
}

export interface events_events_items_payload_peerState {
  __typename: "EventPeerState";
  peerId: string;
  status: PeerStatus | null;
  tracks: events_events_items_payload_peerState_tracks[];
}

export interface events_events_items_payload_message {
  __typename: "EventMessage";
  fromId: string;
  text: string;
}

export interface events_events_items_payload_recording {
  __typename: "EventRecording";
  status: RecordingStatus;
}

export interface events_events_items_payload_trackRecording {
  __typename: "EventTrackRecording";
  id: string;
  trackId: string;
}

export interface events_events_items_payload {
  __typename: "EventPayload";
  peerState: events_events_items_payload_peerState | null;
  message: events_events_items_payload_message | null;
  recording: events_events_items_payload_recording | null;
  trackRecording: events_events_items_payload_trackRecording | null;
}

export interface events_events_items {
  __typename: "Event";
  id: string;
  roomId: string;
  createdAt: number;
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

export interface profiles_profiles {
  __typename: "Profiles";
  items: profiles_profiles_items[];
  nextFrom: string;
}

export interface profiles {
  profiles: profiles_profiles;
}

export interface profilesVariables {
  where: ProfileLookup;
  limit: number;
  from?: string | null;
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
  request: ProfileMask;
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
  state: TalkState;
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

export enum EventOrder {
  ASC = "ASC",
  DESC = "DESC",
}

export enum Hint {
  camera = "camera",
  device_audio = "device_audio",
  screen = "screen",
  user_audio = "user_audio",
}

export enum PeerStatus {
  joined = "joined",
  left = "left",
}

export enum RecordingStatus {
  started = "started",
  stopped = "stopped",
}

export enum TalkState {
  CREATED = "CREATED",
  ENDED = "ENDED",
  LIVE = "LIVE",
  RECORDING = "RECORDING",
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
  createdAt: string;
  id: string;
}

export interface EventLimit {
  count: number;
  seconds?: number | null;
}

export interface EventLookup {
  roomId: string;
}

export interface ProfileLookup {
  ownerIds?: string[] | null;
  handle?: string | null;
}

export interface ProfileMask {
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
