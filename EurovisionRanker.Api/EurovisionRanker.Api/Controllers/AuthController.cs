using Microsoft.AspNetCore.Mvc;
using EurovisionRanker.Api.Models;
using Microsoft.AspNetCore.Authorization;

[ApiController]
[Route("api/[controller]")]
public class AuthController : ControllerBase
{
    private readonly IAuthService _authService;

    public AuthController(IAuthService authService)
    {
        _authService = authService;
    }

    [HttpPost("register")]
    public async Task<IActionResult> Register(RegisterRequest req)
    {
        var userId = await _authService.RegisterAsync(req);

        if (userId == null)
        {
            return BadRequest("Username already taken.");
        }

        return Ok(new { UserId = userId });
    }

    [HttpPost("login")]
    [AllowAnonymous]
    public async Task<IActionResult> Login([FromBody] LoginRequest request)
    {
        var response = await _authService.LoginAsync(request.Username, request.Password);

        if (response == null)
        {
            return Unauthorized("Invalid credentials.");
        }

        return Ok(response);
    }
}