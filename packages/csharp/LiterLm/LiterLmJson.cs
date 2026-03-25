using System.Text.Json;
using System.Text.Json.Serialization;

namespace LiterLm;

/// <summary>
/// Shared <see cref="JsonSerializerOptions"/> for all liter-lm serialization operations.
/// </summary>
/// <remarks>
/// Uses <see cref="JsonSerializerDefaults.Web"/> as the baseline (camelCase, tolerant
/// deserialization) and adds the custom converters required by the liter-lm type hierarchy.
/// </remarks>
public static class LiterLmJson
{
    private static readonly JsonSerializerOptions _options = BuildOptions();

    /// <summary>
    /// Gets the shared <see cref="JsonSerializerOptions"/> instance configured for
    /// the liter-lm wire format.
    /// </summary>
    public static JsonSerializerOptions SerializerOptions => _options;

    private static JsonSerializerOptions BuildOptions()
    {
        var options = new JsonSerializerOptions(JsonSerializerDefaults.Web)
        {
            DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull,
            PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
            WriteIndented = false,
        };
        return options;
    }

    /// <summary>
    /// Serializes <paramref name="value"/> to a JSON string using the liter-lm options.
    /// </summary>
    public static string Serialize<T>(T value) =>
        JsonSerializer.Serialize(value, _options);

    /// <summary>
    /// Deserializes <paramref name="json"/> to <typeparamref name="T"/> using the liter-lm options.
    /// </summary>
    public static T? Deserialize<T>(string json) =>
        JsonSerializer.Deserialize<T>(json, _options);
}
