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
    reaction?:    Reaction;
}

export interface RoomEvent {
    createdAt: number;
    id:        string;
    payload:   EventPayload;
    roomId:    string;
}

export interface EventPayload {
    message?:     EventMessage;
    peerState?:   EventPeerState;
    reaction?:    EventReaction;
    recording?:   EventRecording;
    trackRecord?: EventTrackRecord;
}

export interface EventMessage {
    fromId: string;
    text:   string;
}

export interface EventPeerState {
    peerId:    string;
    sessionId: string;
    status?:   PeerStatus;
}

export enum PeerStatus {
    Joined = "JOINED",
    Left = "LEFT",
}

export interface EventReaction {
    fromId:   string;
    reaction: Reaction;
}

export interface Reaction {
    clap?: ReactionClap;
}

export interface ReactionClap {
    isStarting: boolean;
}

export interface EventRecording {
    status: RecordingEventStatus;
}

export enum RecordingEventStatus {
    Started = "STARTED",
    Stopped = "STOPPED",
}

export interface EventTrackRecord {
    kind:     TrackKind;
    recordId: string;
    source:   TrackSource;
}

export enum TrackKind {
    Audio = "AUDIO",
    Video = "VIDEO",
}

export enum TrackSource {
    Camera = "CAMERA",
    Microphone = "MICROPHONE",
    Screen = "SCREEN",
    ScreenAudio = "SCREEN_AUDIO",
    Unknown = "UNKNOWN",
}

export interface PeerMessage {
    text: string;
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
        { json: "reaction", js: "reaction", typ: u(undefined, r("Reaction")) },
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
        { json: "reaction", js: "reaction", typ: u(undefined, r("EventReaction")) },
        { json: "recording", js: "recording", typ: u(undefined, r("EventRecording")) },
        { json: "trackRecord", js: "trackRecord", typ: u(undefined, r("EventTrackRecord")) },
    ], false),
    "EventMessage": o([
        { json: "fromId", js: "fromId", typ: "" },
        { json: "text", js: "text", typ: "" },
    ], false),
    "EventPeerState": o([
        { json: "peerId", js: "peerId", typ: "" },
        { json: "sessionId", js: "sessionId", typ: "" },
        { json: "status", js: "status", typ: u(undefined, r("PeerStatus")) },
    ], false),
    "EventReaction": o([
        { json: "fromId", js: "fromId", typ: "" },
        { json: "reaction", js: "reaction", typ: r("Reaction") },
    ], false),
    "Reaction": o([
        { json: "clap", js: "clap", typ: u(undefined, r("ReactionClap")) },
    ], false),
    "ReactionClap": o([
        { json: "isStarting", js: "isStarting", typ: true },
    ], false),
    "EventRecording": o([
        { json: "status", js: "status", typ: r("RecordingEventStatus") },
    ], false),
    "EventTrackRecord": o([
        { json: "kind", js: "kind", typ: r("TrackKind") },
        { json: "recordId", js: "recordId", typ: "" },
        { json: "source", js: "source", typ: r("TrackSource") },
    ], false),
    "PeerMessage": o([
        { json: "text", js: "text", typ: "" },
    ], false),
    "PeerStatus": [
        "JOINED",
        "LEFT",
    ],
    "RecordingEventStatus": [
        "STARTED",
        "STOPPED",
    ],
    "TrackKind": [
        "AUDIO",
        "VIDEO",
    ],
    "TrackSource": [
        "CAMERA",
        "MICROPHONE",
        "SCREEN",
        "SCREEN_AUDIO",
        "UNKNOWN",
    ],
};
