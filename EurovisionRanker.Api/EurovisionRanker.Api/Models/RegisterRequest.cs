namespace EurovisionRanker.Api.Models
{
    public class RegisterRequest
    {
        public string Username { get; set; }
        public string Password { get; set; }
        public string Email { get; set; }
        public string ProfilePicUrl { get; set; }
        public string Description { get; set; }
    }
}
