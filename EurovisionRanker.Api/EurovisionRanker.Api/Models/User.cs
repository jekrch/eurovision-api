namespace EurovisionRanker.Api.Models;

public class User
{
    public Guid Id { get; set; }

    public required string Username { get; set; }

    public required string Email { get; set; }

    public required string PasswordHash { get; set; }

    public string? ProfilePicUrl { get; set; }

    public string? Description { get; set; }

    public DateTime CreatedAt { get; set; }
}