using System.ComponentModel.DataAnnotations;

namespace EurovisionRanker.Api.Models
{
    public class RankingResponse
    {
        public required Guid Id { get; init; }

        [Required]
        public required string Name { get; init; }

        public string? Description { get; init; }

        [Range(1956, 2100)]
        public required int Year { get; init; }

        [Required]
        public required string RankingString { get; init; }
    }
}
