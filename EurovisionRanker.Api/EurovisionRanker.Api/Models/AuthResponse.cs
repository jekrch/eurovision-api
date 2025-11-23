namespace EurovisionRanker.Api.Models
{
    public class AuthResponse
    {
        public required string Token { get; set; }
        public required string Username { get; set; }
        public required Guid UserId { get; set; }
    }
}
