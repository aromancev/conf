module.exports = {
  client: {
    service: {
      name: "confa",
      localSchemaFile: ".proto/schema.graphql",
    },
    includes: ["src/**/*.vue", "src/**/*.ts", "src/**/*.graphql"],
  },
}
