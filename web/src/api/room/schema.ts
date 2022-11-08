// To parse this data:
//
//   import { Convert, Message } from "./file";
//
//   const message = Convert.toMessage(json);
//
// These functions will throw an error if the JSON doesn't
// match the expected interface, even if the JSON is valid.

export interface Message {
    payload:     MessagePayload;
    requestId?:  string;
    responseId?: string;
}

export interface MessagePayload {
    event?:       RoomEvent;
    peerMessage?: PeerMessage;
    signal?:      Signal;
    state?:       PeerState;
}

export interface RoomEvent {
    createdAt: number;
    id:        string;
    payload:   EventPayload;
    roomId:    string;
}

export interface EventPayload {
    message?:        EventMessage;
    peerState?:      EventPeerState;
    recording?:      EventRecording;
    trackRecording?: EventTrackRecording;
}

export interface EventMessage {
    fromId: string;
    text:   string;
}

export interface EventPeerState {
    peerId:    string;
    sessionId: string;
    status?:   PeerStatus;
    tracks?:   Track[];
}

export enum PeerStatus {
    Joined = "joined",
    Left = "left",
}

export interface Track {
    hint: Hint;
    id:   string;
}

export enum Hint {
    Camera = "camera",
    DeviceAudio = "device_audio",
    Screen = "screen",
    UserAudio = "user_audio",
}

export interface EventRecording {
    status: RecordingEventStatus;
}

export enum RecordingEventStatus {
    Started = "started",
    Stopped = "stopped",
}

export interface EventTrackRecording {
    id:      string;
    trackId: string;
}

export interface PeerMessage {
    text: string;
}

export interface Signal {
    answer?:  SignalAnswer;
    join?:    SignalJoin;
    offer?:   SignalOffer;
    trickle?: SignalTrickle;
}

export interface SignalAnswer {
    description: SessionDescription;
}

export interface SessionDescription {
    sdp:  string;
    type: SDPType;
}

export enum SDPType {
    Answer = "answer",
    Offer = "offer",
    Pranswer = "pranswer",
    Rollback = "rollback",
}

export interface SignalJoin {
    description: SessionDescription;
    sessionId:   string;
    userId:      string;
}

export interface SignalOffer {
    description: SessionDescription;
}

export interface SignalTrickle {
    candidate: ICECandidateInit;
    target:    number;
}

export interface ICECandidateInit {
    candidate:         string;
    sdpMid?:           string;
    sdpMLineIndex?:    number;
    usernameFragment?: string;
}

export interface PeerState {
    tracks?: Track[];
}

// Converts JSON strings to/from your types
// and asserts the results of JSON.parse at runtime
export class Convert {
    public static toMessage(json: string): Message {
        return cast(JSON.parse(json), r("Message"));
    }

    public static messageToJson(value: Message): string {
        return JSON.stringify(uncast(value, r("Message")), null, 2);
    }
}

function invalidValue(typ: any, val: any, key: any = ''): never {
    if (key) {
        throw Error(`Invalid value for key "${key}". Expected type ${JSON.stringify(typ)} but got ${JSON.stringify(val)}`);
    }
    throw Error(`Invalid value ${JSON.stringify(val)} for type ${JSON.stringify(typ)}`, );
}

function jsonToJSProps(typ: any): any {
    if (typ.jsonToJS === undefined) {
        const map: any = {};
        typ.props.forEach((p: any) => map[p.json] = { key: p.js, typ: p.typ });
        typ.jsonToJS = map;
    }
    return typ.jsonToJS;
}

function jsToJSONProps(typ: any): any {
    if (typ.jsToJSON === undefined) {
        const map: any = {};
        typ.props.forEach((p: any) => map[p.js] = { key: p.json, typ: p.typ });
        typ.jsToJSON = map;
    }
    return typ.jsToJSON;
}

function transform(val: any, typ: any, getProps: any, key: any = ''): any {
    function transformPrimitive(typ: string, val: any): any {
        if (typeof typ === typeof val) return val;
        return invalidValue(typ, val, key);
    }

    function transformUnion(typs: any[], val: any): any {
        // val must validate against one typ in typs
        const l = typs.length;
        for (let i = 0; i < l; i++) {
            const typ = typs[i];
            try {
                return transform(val, typ, getProps);
            } catch (_) {}
        }
        return invalidValue(typs, val);
    }

    function transformEnum(cases: string[], val: any): any {
        if (cases.indexOf(val) !== -1) return val;
        return invalidValue(cases, val);
    }

    function transformArray(typ: any, val: any): any {
        // val must be an array with no invalid elements
        if (!Array.isArray(val)) return invalidValue("array", val);
        return val.map(el => transform(el, typ, getProps));
    }

    function transformDate(val: any): any {
        if (val === null) {
            return null;
        }
        const d = new Date(val);
        if (isNaN(d.valueOf())) {
            return invalidValue("Date", val);
        }
        return d;
    }

    function transformObject(props: { [k: string]: any }, additional: any, val: any): any {
        if (val === null || typeof val !== "object" || Array.isArray(val)) {
            return invalidValue("object", val);
        }
        const result: any = {};
        Object.getOwnPropertyNames(props).forEach(key => {
            const prop = props[key];
            const v = Object.prototype.hasOwnProperty.call(val, key) ? val[key] : undefined;
            result[prop.key] = transform(v, prop.typ, getProps, prop.key);
        });
        Object.getOwnPropertyNames(val).forEach(key => {
            if (!Object.prototype.hasOwnProperty.call(props, key)) {
                result[key] = transform(val[key], additional, getProps, key);
            }
        });
        return result;
    }

    if (typ === "any") return val;
    if (typ === null) {
        if (val === null) return val;
        return invalidValue(typ, val);
    }
    if (typ === false) return invalidValue(typ, val);
    while (typeof typ === "object" && typ.ref !== undefined) {
        typ = typeMap[typ.ref];
    }
    if (Array.isArray(typ)) return transformEnum(typ, val);
    if (typeof typ === "object") {
        return typ.hasOwnProperty("unionMembers") ? transformUnion(typ.unionMembers, val)
            : typ.hasOwnProperty("arrayItems")    ? transformArray(typ.arrayItems, val)
            : typ.hasOwnProperty("props")         ? transformObject(getProps(typ), typ.additional, val)
            : invalidValue(typ, val);
    }
    // Numbers can be parsed by Date but shouldn't be.
    if (typ === Date && typeof val !== "number") return transformDate(val);
    return transformPrimitive(typ, val);
}

function cast<T>(val: any, typ: any): T {
    return transform(val, typ, jsonToJSProps);
}

function uncast<T>(val: T, typ: any): any {
    return transform(val, typ, jsToJSONProps);
}

function a(typ: any) {
    return { arrayItems: typ };
}

function u(...typs: any[]) {
    return { unionMembers: typs };
}

function o(props: any[], additional: any) {
    return { props, additional };
}

function m(additional: any) {
    return { props: [], additional };
}

function r(name: string) {
    return { ref: name };
}

const typeMap: any = {
    "Message": o([
        { json: "payload", js: "payload", typ: r("MessagePayload") },
        { json: "requestId", js: "requestId", typ: u(undefined, "") },
        { json: "responseId", js: "responseId", typ: u(undefined, "") },
    ], false),
    "MessagePayload": o([
        { json: "event", js: "event", typ: u(undefined, r("RoomEvent")) },
        { json: "peerMessage", js: "peerMessage", typ: u(undefined, r("PeerMessage")) },
        { json: "signal", js: "signal", typ: u(undefined, r("Signal")) },
        { json: "state", js: "state", typ: u(undefined, r("PeerState")) },
    ], false),
    "RoomEvent": o([
        { json: "createdAt", js: "createdAt", typ: 3.14 },
        { json: "id", js: "id", typ: "" },
        { json: "payload", js: "payload", typ: r("EventPayload") },
        { json: "roomId", js: "roomId", typ: "" },
    ], false),
    "EventPayload": o([
        { json: "message", js: "message", typ: u(undefined, r("EventMessage")) },
        { json: "peerState", js: "peerState", typ: u(undefined, r("EventPeerState")) },
        { json: "recording", js: "recording", typ: u(undefined, r("EventRecording")) },
        { json: "trackRecording", js: "trackRecording", typ: u(undefined, r("EventTrackRecording")) },
    ], false),
    "EventMessage": o([
        { json: "fromId", js: "fromId", typ: "" },
        { json: "text", js: "text", typ: "" },
    ], false),
    "EventPeerState": o([
        { json: "peerId", js: "peerId", typ: "" },
        { json: "sessionId", js: "sessionId", typ: "" },
        { json: "status", js: "status", typ: u(undefined, r("PeerStatus")) },
        { json: "tracks", js: "tracks", typ: u(undefined, a(r("Track"))) },
    ], false),
    "Track": o([
        { json: "hint", js: "hint", typ: r("Hint") },
        { json: "id", js: "id", typ: "" },
    ], false),
    "EventRecording": o([
        { json: "status", js: "status", typ: r("RecordingEventStatus") },
    ], false),
    "EventTrackRecording": o([
        { json: "id", js: "id", typ: "" },
        { json: "trackId", js: "trackId", typ: "" },
    ], false),
    "PeerMessage": o([
        { json: "text", js: "text", typ: "" },
    ], false),
    "Signal": o([
        { json: "answer", js: "answer", typ: u(undefined, r("SignalAnswer")) },
        { json: "join", js: "join", typ: u(undefined, r("SignalJoin")) },
        { json: "offer", js: "offer", typ: u(undefined, r("SignalOffer")) },
        { json: "trickle", js: "trickle", typ: u(undefined, r("SignalTrickle")) },
    ], false),
    "SignalAnswer": o([
        { json: "description", js: "description", typ: r("SessionDescription") },
    ], false),
    "SessionDescription": o([
        { json: "sdp", js: "sdp", typ: "" },
        { json: "type", js: "type", typ: r("SDPType") },
    ], false),
    "SignalJoin": o([
        { json: "description", js: "description", typ: r("SessionDescription") },
        { json: "sessionId", js: "sessionId", typ: "" },
        { json: "userId", js: "userId", typ: "" },
    ], false),
    "SignalOffer": o([
        { json: "description", js: "description", typ: r("SessionDescription") },
    ], false),
    "SignalTrickle": o([
        { json: "candidate", js: "candidate", typ: r("ICECandidateInit") },
        { json: "target", js: "target", typ: 0 },
    ], false),
    "ICECandidateInit": o([
        { json: "candidate", js: "candidate", typ: "" },
        { json: "sdpMid", js: "sdpMid", typ: u(undefined, "") },
        { json: "sdpMLineIndex", js: "sdpMLineIndex", typ: u(undefined, 0) },
        { json: "usernameFragment", js: "usernameFragment", typ: u(undefined, "") },
    ], false),
    "PeerState": o([
        { json: "tracks", js: "tracks", typ: u(undefined, a(r("Track"))) },
    ], false),
    "PeerStatus": [
        "joined",
        "left",
    ],
    "Hint": [
        "camera",
        "device_audio",
        "screen",
        "user_audio",
    ],
    "RecordingEventStatus": [
        "started",
        "stopped",
    ],
    "SDPType": [
        "answer",
        "offer",
        "pranswer",
        "rollback",
    ],
};
