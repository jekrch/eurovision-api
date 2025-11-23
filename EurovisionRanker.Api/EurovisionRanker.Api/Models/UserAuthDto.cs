namespace EurovisionRanker.Api.Models
{
    public class UserAuthDto
    {
        public Guid Id { get; set; }
        public string PasswordHash { get; set; }
    }
}
