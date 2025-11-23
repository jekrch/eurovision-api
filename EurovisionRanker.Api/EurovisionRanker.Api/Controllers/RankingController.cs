using EurovisionRanker.Api.Extensions; 
using EurovisionRanker.Api.Models;
using EurovisionRanker.Api.Repositories;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;

namespace EurovisionRanker.Api.Controllers;

[ApiController]
[Route("api/[controller]")]
[Authorize]
public class RankingController : ControllerBase
{
    private readonly IRankingRepository _repository;

    public RankingController(IRankingRepository repository)
    {
        _repository = repository;
    }

    [HttpPost]
    public async Task<IActionResult> Create([FromBody] CreateRankingRequest request)
    {
        // Usage of the clean extension method
        var userId = User.GetUserId();
        var created = await _repository.CreateAsync(userId, request);

        return CreatedAtAction(nameof(GetById), new { id = created.Id }, created);
    }

    [HttpGet]
    public async Task<IActionResult> GetMyRankings()
    {
        var userId = User.GetUserId();
        var rankings = await _repository.GetAllForUserAsync(userId);
        return Ok(rankings);
    }

    [HttpGet("{id:guid}")]
    public async Task<IActionResult> GetById(Guid id)
    {
        var ranking = await _repository.GetByIdAsync(id);
        if (ranking == null) return NotFound();

        return Ok(ranking);
    }

    [HttpPut("{id:guid}")]
    public async Task<IActionResult> Update(Guid id, [FromBody] UpdateRankingRequest request)
    {
        var userId = User.GetUserId();

        // The Repository handles the ownership check via the SQL WHERE clause
        var success = await _repository.UpdateAsync(id, userId, request);

        if (!success) return NotFound("Ranking not found or access denied.");

        return NoContent();
    }

    [HttpDelete("{id:guid}")]
    public async Task<IActionResult> Delete(Guid id)
    {
        var userId = User.GetUserId();
        var success = await _repository.DeleteAsync(id, userId);

        if (!success) return NotFound("Ranking not found or access denied.");

        return NoContent();
    }
}