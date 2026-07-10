using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using {{ProjectName}}.Infrastructure.Persistence;

namespace {{ProjectName}}.Infrastructure;

public static class DependencyInjection
{
    public static IServiceCollection AddInfrastructure(
        this IServiceCollection services,
        ConfigurationManager configuration)
    {
        services.AddDbContext<AppDbContext>(options =>
            options.UseNpgsql(
                configuration.GetConnectionString("AppDb"),
                npgsqlOptions =>
                {
                    // Register PostgreSQL enum mappings here.
                }));

        return services;
    }

    public static IServiceCollection AddServices(
        this IServiceCollection services)
    {
        return services;
    }
}