import type { Schema, Struct } from '@strapi/strapi';

export interface TestTest extends Struct.ComponentSchema {
  collectionName: 'components_test_tests';
  info: {
    displayName: 'test';
    icon: 'calendar';
  };
  attributes: {
    mail: Schema.Attribute.Email;
  };
}

declare module '@strapi/strapi' {
  export module Public {
    export interface ComponentSchemas {
      'test.test': TestTest;
    }
  }
}
