using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using Microsoft.IdentityModel.Tokens;

namespace EurovisionRanker.Api.Extensions;

public static class ClaimsPrincipalExtensions
{
    public static Guid GetUserId(this ClaimsPrincipal user)
    {
        var idClaim = user.FindFirst(JwtRegisteredClaimNames.Sub)
                      ?? user.FindFirst(ClaimTypes.NameIdentifier)
                      ?? user.FindFirst("sub"); // Fallback for some JWT providers

        if (idClaim != null && Guid.TryParse(idClaim.Value, out var userId))
        {
            return userId;
        }

        throw new UnauthorizedAccessException("Invalid Token: User ID claim is missing or invalid.");
    }
}