# STAGE 1: Build
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src

COPY ["EurovisionRanker.Api/EurovisionRanker.Api.csproj", "EurovisionRanker.Api/"]
RUN dotnet restore "EurovisionRanker.Api/EurovisionRanker.Api.csproj"

COPY . .
WORKDIR "/src/EurovisionRanker.Api"

# Build and Publish
RUN dotnet publish "EurovisionRanker.Api.csproj" -c Release -o /app/publish /p:UseAppHost=false

# STAGE 2: Run
FROM mcr.microsoft.com/dotnet/aspnet:10.0 AS final
WORKDIR /app
EXPOSE 8080
COPY --from=build /app/publish .
ENTRYPOINT ["dotnet", "EurovisionRanker.Api.dll"]