namespace EurovisionRanker.Api.Models
{
    public class Ranking
    {
        public Guid Id { get; set; }
        public Guid UserId { get; set; }
        public required string Name { get; set; }
        public string? Description { get; set; }
        public int Year { get; set; }

        // The raw URL-encoded string 
        public required string RankingString { get; set; }

        public DateTime UpdatedAt { get; set; }
    }
}
