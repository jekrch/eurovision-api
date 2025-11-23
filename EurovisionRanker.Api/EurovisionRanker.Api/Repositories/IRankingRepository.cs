using Dapper;
using EurovisionRanker.Api.Models;
using Npgsql;

namespace EurovisionRanker.Api.Repositories;

public interface IRankingRepository
{
    Task<Ranking> CreateAsync(Guid userId, CreateRankingRequest request);
    Task<IEnumerable<Ranking>> GetAllForUserAsync(Guid userId);
    Task<Ranking?> GetByIdAsync(Guid id);
    Task<bool> UpdateAsync(Guid id, Guid userId, UpdateRankingRequest request);
    Task<bool> DeleteAsync(Guid id, Guid userId);
}

public class RankingRepository : IRankingRepository
{
    private readonly NpgsqlDataSource _dataSource;

    public RankingRepository(NpgsqlDataSource dataSource)
    {
        _dataSource = dataSource;
    }

    public async Task<Ranking> CreateAsync(Guid userId, CreateRankingRequest req)
    {
        const string sql = @"
            INSERT INTO ranker.ranking (user_id, name, description, year, ranking_string)
            VALUES (@UserId, @Name, @Description, @Year, @RankingString)
            RETURNING 
                id, 
                user_id as UserId, 
                name, 
                description, 
                year, 
                ranking_string as RankingString, 
                updated_at as UpdatedAt";

        using var conn = await _dataSource.OpenConnectionAsync();
        return await conn.QuerySingleAsync<Ranking>(sql, new
        {
            UserId = userId,
            req.Name,
            req.Description,
            req.Year,
            req.RankingString
        });
    }

    public async Task<IEnumerable<Ranking>> GetAllForUserAsync(Guid userId)
    {
        const string sql = @"
            SELECT 
                id, 
                user_id as UserId, 
                name, 
                description, 
                year, 
                ranking_string as RankingString, 
                updated_at as UpdatedAt
            FROM ranker.ranking
            WHERE user_id = @UserId
            ORDER BY updated_at DESC";

        using var conn = await _dataSource.OpenConnectionAsync();
        return await conn.QueryAsync<Ranking>(sql, new { UserId = userId });
    }

    public async Task<Ranking?> GetByIdAsync(Guid id)
    {
        const string sql = @"
            SELECT 
                id, 
                user_id as UserId, 
                name, 
                description, 
                year, 
                ranking_string as RankingString, 
                updated_at as UpdatedAt
            FROM ranker.ranking
            WHERE id = @Id";

        using var conn = await _dataSource.OpenConnectionAsync();
        return await conn.QuerySingleOrDefaultAsync<Ranking>(sql, new { Id = id });
    }

    public async Task<bool> UpdateAsync(Guid id, Guid userId, UpdateRankingRequest req)
    {
        // We include 'user_id = @UserId' in the WHERE clause.
        // This ensures a user cannot update someone else's ranking.
        const string sql = @"
            UPDATE ranker.ranking
            SET 
                name = @Name,
                description = @Description,
                ranking_string = @RankingString,
                updated_at = NOW()
            WHERE id = @Id AND user_id = @UserId";

        using var conn = await _dataSource.OpenConnectionAsync();
        var rowsAffected = await conn.ExecuteAsync(sql, new
        {
            Id = id,
            UserId = userId,
            req.Name,
            req.Description,
            req.RankingString
        });

        return rowsAffected > 0;
    }

    public async Task<bool> DeleteAsync(Guid id, Guid userId)
    {
        const string sql = "DELETE FROM ranker.ranking WHERE id = @Id AND user_id = @UserId";

        using var conn = await _dataSource.OpenConnectionAsync();
        var rowsAffected = await conn.ExecuteAsync(sql, new { Id = id, UserId = userId });

        return rowsAffected > 0;
    }
}