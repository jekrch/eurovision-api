using Dapper;
using Npgsql;
using EurovisionRanker.Api.Models;

public interface IUserRepository
{
    Task<bool> ExistsAsync(string username);
    Task<Guid> CreateAsync(User user);
    Task<User?> GetByUsernameAsync(string username);
}

public class UserRepository : IUserRepository
{
    private readonly NpgsqlDataSource _dataSource;

    public UserRepository(NpgsqlDataSource dataSource)
    {
        _dataSource = dataSource;
    }

    public async Task<bool> ExistsAsync(string username)
    {
        const string sql = "SELECT EXISTS(SELECT 1 FROM ranker.users WHERE username = @Username)";

        using var conn = await _dataSource.OpenConnectionAsync();
        return await conn.ExecuteScalarAsync<bool>(sql, new { Username = username });
    }

    public async Task<Guid> CreateAsync(User user)
    {
        const string sql = @"
            INSERT INTO ranker.users (username, email, password_hash, profile_pic_url, description) 
            VALUES (@Username, @Email, @PasswordHash, @ProfilePicUrl, @Description)
            RETURNING id";

        using var conn = await _dataSource.OpenConnectionAsync();
        return await conn.ExecuteScalarAsync<Guid>(sql, user);
    }

    public async Task<User?> GetByUsernameAsync(string username)
    {
        // Explicitly selecting columns is more memory efficient than SELECT *
        const string sql = @"
            SELECT id, username, email, password_hash as PasswordHash, profile_pic_url as ProfilePicUrl, description 
            FROM ranker.users 
            WHERE username = @Username";

        using var conn = await _dataSource.OpenConnectionAsync();
        return await conn.QuerySingleOrDefaultAsync<User>(sql, new { Username = username });
    }
}