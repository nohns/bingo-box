local claims = {
  email_verified: false
} + std.extVar('claims');

{
  identity: {
    traits: {
      // Allowing unverified email addresses enables account
      // enumeration attacks, especially if the value is used for
      // e.g. verification or as a password login identifier
      email: claims.email,
      name: {
        first: claims.given_name,
        last: claims.family_name,
      },
    },
  },
}