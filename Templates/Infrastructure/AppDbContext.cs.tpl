using Microsoft.EntityFrameworkCore;

namespace {{ProjectName}}.Infrastructure.Persistence;

public class AppDbContext : DbContext
{
    public AppDbContext(
        DbContextOptions<AppDbContext> options)
        : base(options)
    {
    }
}