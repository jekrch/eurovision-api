namespace EurovisionRanker.Api.Models
{
    public class UpdateRankingRequest
    {
        public required string Name { get; set; }
        public string? Description { get; set; }
        public required string RankingString { get; set; }
    }
}
