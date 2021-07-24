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
  from: string;
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
}

export interface createConfa {
  createConfa: createConfa_createConfa;
}

/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

//==============================================================
// START Enums and Input Objects
//==============================================================

export interface ConfaInput {
  id?: string | null;
  ownerId?: string | null;
  handle?: string | null;
}

//==============================================================
// END Enums and Input Objects
//==============================================================
