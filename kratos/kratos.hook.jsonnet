function(ctx) {
  key: "<Your-Api-Key>",
  message: {
    from_email: "hello@example.com",
    subject: "Hello from Ory Kratos",
    text: "Welcome to Ory Kratos",
    to: [
      {
        email: ctx.identity.id,
        type: "to"
      }
    ]
  }
}