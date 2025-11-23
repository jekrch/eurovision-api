namespace EurovisionRanker.Api.Models
{
    public class CreateRankingRequest
    {
        public required string Name { get; set; }
        public string? Description { get; set; }
        public required int Year { get; set; }
        public required string RankingString { get; set; }
    }
}
