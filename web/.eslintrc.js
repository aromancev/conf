module.exports = {
    root: true,
    env: {
        node: true,
        "vue/setup-compiler-macros": true,
    },
    parser: "vue-eslint-parser",
    parserOptions: {
        parser: "@typescript-eslint/parser",
        ecmaVersion: 2020,
        sourceType: "module",
    },
    extends: [
        "plugin:@typescript-eslint/recommended",
        "eslint:recommended",
        "plugin:prettier/recommended",
        "plugin:vue/vue3-recommended",
        "prettier",
    ],
    plugins: ["prettier", "@typescript-eslint"],
    rules: {
        "no-unused-vars": "off",
    },
    globals: {
        "RTCSessionDescriptionInit": "readonly",
        "RTCIceServer": "readonly",
        "RTCRtpEncodingParameters": "readonly"
    },
    "ignorePatterns": ["dist"],
}
