defmodule LiterLm.TypesTest do
  use ExUnit.Case, async: true

  alias LiterLm.Types

  # ─── Message Constructors ─────────────────────────────────────────────────

  describe "system_message/2" do
    test "creates a system message with content" do
      msg = Types.system_message("You are helpful.")

      assert msg.role == "system"
      assert msg.content == "You are helpful."
      refute Map.has_key?(msg, :name)
    end

    test "includes optional name when provided" do
      msg = Types.system_message("Instructions.", name: "sys")

      assert msg.name == "sys"
    end
  end

  describe "user_message/2" do
    test "creates a user message with string content" do
      msg = Types.user_message("Hello!")

      assert msg.role == "user"
      assert msg.content == "Hello!"
    end

    test "creates a user message with multi-part content" do
      parts = [
        %{type: "text", text: "What's in this image?"},
        %{type: "image_url", image_url: %{url: "https://example.com/img.png"}}
      ]

      msg = Types.user_message(parts)

      assert msg.role == "user"
      assert is_list(msg.content)
      assert length(msg.content) == 2
    end
  end

  describe "assistant_message/1" do
    test "creates an assistant message with content" do
      msg = Types.assistant_message(content: "I can help!")

      assert msg.role == "assistant"
      assert msg.content == "I can help!"
    end

    test "creates an assistant message with tool calls" do
      tool_call = %{
        id: "call-1",
        type: "function",
        function: %{name: "get_weather", arguments: "{\"city\": \"Berlin\"}"}
      }

      msg = Types.assistant_message(tool_calls: [tool_call])

      assert msg.role == "assistant"
      assert length(msg.tool_calls) == 1
      assert hd(msg.tool_calls).function.name == "get_weather"
    end

    test "creates an empty assistant message" do
      msg = Types.assistant_message()

      assert msg.role == "assistant"
      refute Map.has_key?(msg, :content)
    end
  end

  describe "tool_message/3" do
    test "creates a tool result message" do
      msg = Types.tool_message("sunny, 22°C", "call-123")

      assert msg.role == "tool"
      assert msg.content == "sunny, 22°C"
      assert msg.tool_call_id == "call-123"
    end
  end

  describe "developer_message/2" do
    test "creates a developer message" do
      msg = Types.developer_message("Focus on accuracy.")

      assert msg.role == "developer"
      assert msg.content == "Focus on accuracy."
    end
  end

  # ─── Tool Constructor ─────────────────────────────────────────────────────

  describe "tool/4" do
    test "creates a function tool" do
      schema = %{"type" => "object", "properties" => %{"city" => %{"type" => "string"}}}
      tool = Types.tool("get_weather", "Get current weather", schema)

      assert tool.type == "function"
      assert tool.function.name == "get_weather"
      assert tool.function.description == "Get current weather"
      assert tool.function.parameters == schema
    end

    test "creates a minimal tool without description or parameters" do
      tool = Types.tool("noop")

      assert tool.type == "function"
      assert tool.function.name == "noop"
      refute Map.has_key?(tool.function, :description)
      refute Map.has_key?(tool.function, :parameters)
    end
  end

  # ─── Finish Reason Parsing ────────────────────────────────────────────────

  describe "parse_finish_reason/1" do
    test "maps known finish reasons" do
      assert Types.parse_finish_reason("stop") == :stop
      assert Types.parse_finish_reason("length") == :length
      assert Types.parse_finish_reason("tool_calls") == :tool_calls
      assert Types.parse_finish_reason("content_filter") == :content_filter
      assert Types.parse_finish_reason("function_call") == :function_call
    end

    test "maps nil to nil" do
      assert Types.parse_finish_reason(nil) == nil
    end

    test "maps unknown values to :other" do
      assert Types.parse_finish_reason("some_new_value") == :other
    end
  end

  # ─── JSON Round-Trip via Jason ────────────────────────────────────────────

  describe "JSON serialization" do
    test "system message serializes correctly" do
      msg = Types.system_message("Hello.")
      {:ok, json} = Jason.encode(msg)
      {:ok, decoded} = Jason.decode(json)

      assert decoded["role"] == "system"
      assert decoded["content"] == "Hello."
    end

    test "chat request serializes with snake_case keys" do
      request = %{
        model: "gpt-4o-mini",
        messages: [Types.user_message("Hi")],
        max_tokens: 100,
        temperature: 0.7
      }

      {:ok, json} = Jason.encode(request)

      assert json =~ "\"max_tokens\""
      assert json =~ "\"temperature\""
      refute json =~ "\"stream\""
    end

    test "embedding request serializes input as string" do
      request = %{
        model: "text-embedding-3-small",
        input: "The quick brown fox"
      }

      {:ok, json} = Jason.encode(request)

      assert json =~ "text-embedding-3-small"
      assert json =~ "The quick brown fox"
    end

    test "embedding request serializes input as list" do
      request = %{
        model: "text-embedding-3-small",
        input: ["first", "second"]
      }

      {:ok, json} = Jason.encode(request)

      assert json =~ "\"first\""
      assert json =~ "\"second\""
    end
  end
end
