// import type { Core } from '@strapi/strapi';
// docs : https://strapi.io/blog/extending-and-building-custom-resolvers-with-strapi-v4
export default {
  /**
   * An asynchronous register function that runs before
   * your application is initialized.
   *
   * This gives you an opportunity to extend code.
   */
  register({ strapi }) {
    const extensionService = strapi.service("plugin::graphql.extension");
    extensionService.use(({ strapi }) => ({
      typeDefs: `
        type Query {
          projects: [Project]
        }
      `,
      resolvers: {
        Query: {
          projects: {
            resolve: async (parent, args, context) => {
              console.log("args:",args)
              const data = await strapi.services["api::project.project"].find();

            
              return data.results.map( async (item) =>  {
                const taskData = await strapi.services["api::project-task.project-task"].find({
                  filters: { project: {
                    documentId: item.documentId,
                  }},
                });

                const totalEstimatedTimeMinute = taskData.results.reduce((total, task) => total + task.estimated_time_minute, 0);  
                item.total_estimated_time_minute = totalEstimatedTimeMinute
                item.total_estimated_cost_usd = 200
                return item
              });
            },
          },
        },
      },
      resolversConfig: {
        "Query.projects": {
          auth: false,
        },
      },
    }));
  },

  /**
   * An asynchronous bootstrap function that runs before
   * your application gets started.
   *
   * This gives you an opportunity to set up your data model,
   * run jobs, or perform some special logic.
   */
  bootstrap(/* { strapi }: { strapi: Core.Strapi } */) {},
};
