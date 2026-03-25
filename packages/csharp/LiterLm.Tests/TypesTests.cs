using System.Text.Json;
using FluentAssertions;

namespace LiterLm.Tests;

/// <summary>Unit tests for liter-lm C# type serialization and deserialization.</summary>
public class TypesTests
{
    // ─── Messages ─────────────────────────────────────────────────────────────

    [Fact]
    public void SystemMessage_RoundTrips()
    {
        var msg = new SystemMessage("You are helpful.");
        var json = LiterLmJson.Serialize(msg);

        json.Should().Contain("\"role\":\"system\"");
        json.Should().Contain("You are helpful.");
        json.Should().NotContain("\"name\"");

        var decoded = LiterLmJson.Deserialize<Message>(json);
        decoded.Should().BeOfType<SystemMessage>()
            .Which.Content.Should().Be("You are helpful.");
    }

    [Fact]
    public void UserMessage_TextContent_RoundTrips()
    {
        var msg = new UserMessage("Hello!");
        var json = LiterLmJson.Serialize(msg);

        json.Should().Contain("\"role\":\"user\"");
        json.Should().Contain("Hello!");
    }

    [Fact]
    public void AssistantMessage_WithToolCalls_RoundTrips()
    {
        var toolCall = new ToolCall("call-1", ToolType.Function,
            new FunctionCall("get_weather", "{\"city\": \"Berlin\"}"));
        var msg = new AssistantMessage(ToolCalls: [toolCall]);
        var json = LiterLmJson.Serialize(msg);

        json.Should().Contain("\"role\":\"assistant\"");
        json.Should().Contain("tool_calls");
        json.Should().Contain("get_weather");
    }

    // ─── Chat Request ─────────────────────────────────────────────────────────

    [Fact]
    public void ChatCompletionRequest_OmitsNullFields()
    {
        var request = new ChatCompletionRequest(
            Model: "gpt-4o-mini",
            Messages: [new UserMessage("Hi")],
            MaxTokens: 100,
            Temperature: 0.7);

        var json = LiterLmJson.Serialize(request);

        json.Should().Contain("\"model\":\"gpt-4o-mini\"");
        json.Should().Contain("\"max_tokens\":100");
        json.Should().Contain("\"temperature\":0.7");
        json.Should().NotContain("\"stream\"");
        json.Should().NotContain("\"seed\"");
    }

    // ─── Finish Reason ────────────────────────────────────────────────────────

    [Theory]
    [InlineData("stop", FinishReason.Stop)]
    [InlineData("length", FinishReason.Length)]
    [InlineData("tool_calls", FinishReason.ToolCalls)]
    [InlineData("content_filter", FinishReason.ContentFilter)]
    [InlineData("function_call", FinishReason.FunctionCall)]
    [InlineData("some_new_value", FinishReason.Other)]
    public void FinishReason_Deserializes(string wire, FinishReason expected)
    {
        var json = $"\"{wire}\"";
        var decoded = JsonSerializer.Deserialize<FinishReason>(json, LiterLmJson.SerializerOptions);
        decoded.Should().Be(expected);
    }

    // ─── Usage ────────────────────────────────────────────────────────────────

    [Fact]
    public void Usage_RoundTrips()
    {
        var usage = new Usage(100L, 50L, 150L);
        var json = LiterLmJson.Serialize(usage);

        json.Should().Contain("\"prompt_tokens\":100");
        json.Should().Contain("\"completion_tokens\":50");
        json.Should().Contain("\"total_tokens\":150");

        var decoded = LiterLmJson.Deserialize<Usage>(json);
        decoded.Should().BeEquivalentTo(usage);
    }

    // ─── Tool Choice ──────────────────────────────────────────────────────────

    [Fact]
    public void ToolChoice_Auto_SerializesAsString()
    {
        var json = LiterLmJson.Serialize(ToolChoice.Auto);
        json.Should().Be("\"auto\"");
    }

    [Fact]
    public void ToolChoice_Required_SerializesAsString()
    {
        var json = LiterLmJson.Serialize(ToolChoice.Required);
        json.Should().Be("\"required\"");
    }

    [Fact]
    public void ToolChoice_ForFunction_SerializesCorrectly()
    {
        var choice = ToolChoice.ForFunction("get_weather");
        var json = LiterLmJson.Serialize(choice);

        json.Should().Contain("\"type\":\"function\"");
        json.Should().Contain("\"get_weather\"");
    }

    // ─── Response Format ──────────────────────────────────────────────────────

    [Fact]
    public void ResponseFormat_Text_RoundTrips()
    {
        var json = LiterLmJson.Serialize(ResponseFormat.Text);
        json.Should().Contain("\"type\":\"text\"");
    }

    [Fact]
    public void ResponseFormat_JsonObject_RoundTrips()
    {
        var json = LiterLmJson.Serialize(ResponseFormat.JsonObject);
        json.Should().Contain("\"type\":\"json_object\"");
    }

    // ─── Stop Sequence ────────────────────────────────────────────────────────

    [Fact]
    public void StopSequence_Single_SerializesAsString()
    {
        var stop = StopSequence.FromString("\n\n");
        var json = LiterLmJson.Serialize(stop);
        json.Should().Be("\"\\n\\n\"");
    }

    [Fact]
    public void StopSequence_Multiple_SerializesAsArray()
    {
        var stop = StopSequence.FromList(["stop1", "stop2"]);
        var json = LiterLmJson.Serialize(stop);

        json.Should().Contain("stop1");
        json.Should().Contain("stop2");
        json.Should().StartWith("[");
    }

    // ─── Embedding ────────────────────────────────────────────────────────────

    [Fact]
    public void EmbeddingRequest_SingleInput_RoundTrips()
    {
        var request = new EmbeddingRequest(
            Model: "text-embedding-3-small",
            Input: EmbeddingInput.FromString("Hello, world!"));
        var json = LiterLmJson.Serialize(request);

        json.Should().Contain("\"model\":\"text-embedding-3-small\"");
        json.Should().Contain("Hello, world!");
    }

    [Fact]
    public void EmbeddingRequest_MultipleInput_RoundTrips()
    {
        var request = new EmbeddingRequest(
            Model: "text-embedding-3-small",
            Input: EmbeddingInput.FromList(["first", "second"]));
        var json = LiterLmJson.Serialize(request);

        json.Should().Contain("first");
        json.Should().Contain("second");
    }

    // ─── Models ───────────────────────────────────────────────────────────────

    [Fact]
    public void ModelObject_Deserializes()
    {
        const string json =
            """{"id":"gpt-4o","object":"model","created":1712361441,"owned_by":"openai"}""";
        var model = LiterLmJson.Deserialize<ModelObject>(json);

        model.Should().NotBeNull();
        model!.Id.Should().Be("gpt-4o");
        model.Object.Should().Be("model");
        model.OwnedBy.Should().Be("openai");
    }

    // ─── Errors ───────────────────────────────────────────────────────────────

    [Fact]
    public void InvalidRequestException_HasCorrectErrorCode()
    {
        var ex = new InvalidRequestException("bad field");
        ex.ErrorCode.Should().Be(LlmException.ErrorCodes.InvalidRequest);
        ex.Message.Should().Contain("invalid request");
        ex.Message.Should().Contain("bad field");
    }

    [Fact]
    public void AuthenticationException_HasCorrectErrorCode()
    {
        var ex = new AuthenticationException("invalid key");
        ex.ErrorCode.Should().Be(LlmException.ErrorCodes.Authentication);
    }

    [Fact]
    public void ProviderException_ExposesHttpStatus()
    {
        var ex = new ProviderException(503, "Service Unavailable");
        ex.HttpStatus.Should().Be(503);
        ex.ErrorCode.Should().Be(LlmException.ErrorCodes.ProviderError);
    }
}
