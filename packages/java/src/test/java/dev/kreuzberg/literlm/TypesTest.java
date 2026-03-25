package dev.kreuzberg.literlm;

import static dev.kreuzberg.literlm.Types.*;
import static org.assertj.core.api.Assertions.assertThat;

import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import java.util.List;
import org.assertj.core.api.SoftAssertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

/** Unit tests for liter-lm Java type serialization and deserialization. */
class TypesTest {

	private ObjectMapper mapper;

	@BeforeEach
	void setUp() {
		mapper = new ObjectMapper().configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);
	}

	@Test
	void systemMessage_roundTrips() throws Exception {
		var msg = new SystemMessage("You are helpful.", "sys");
		String json = mapper.writeValueAsString(msg);
		assertThat(json).contains("\"role\":\"system\"");
		assertThat(json).contains("\"content\":\"You are helpful.\"");

		var decoded = mapper.readValue(json, Message.class);
		assertThat(decoded).isInstanceOf(SystemMessage.class);
		assertThat(((SystemMessage) decoded).content()).isEqualTo("You are helpful.");
	}

	@Test
	void userMessage_textOnly_roundTrips() throws Exception {
		var msg = new UserMessage("Hello!");
		String json = mapper.writeValueAsString(msg);
		assertThat(json).contains("\"role\":\"user\"");
		assertThat(json).doesNotContain("\"name\"");

		var decoded = mapper.readValue(json, Message.class);
		assertThat(decoded).isInstanceOf(UserMessage.class);
	}

	@Test
	void assistantMessage_withToolCalls_roundTrips() throws Exception {
		var toolCall = new ToolCall("call-1", ToolType.FUNCTION,
				new FunctionCall("get_weather", "{\"city\": \"Berlin\"}"));
		var msg = new AssistantMessage(List.of(toolCall));
		String json = mapper.writeValueAsString(msg);

		assertThat(json).contains("\"role\":\"assistant\"");
		assertThat(json).contains("\"tool_calls\"");
		assertThat(json).contains("\"get_weather\"");

		var decoded = mapper.readValue(json, Message.class);
		assertThat(decoded).isInstanceOf(AssistantMessage.class);
		var assistant = (AssistantMessage) decoded;
		assertThat(assistant.toolCalls()).hasSize(1);
		assertThat(assistant.toolCalls().getFirst().function().name()).isEqualTo("get_weather");
	}

	@Test
	void chatCompletionRequest_builder_omitsNulls() throws Exception {
		var request = ChatCompletionRequest.builder("gpt-4o-mini", List.of(new UserMessage("Hi"))).maxTokens(100L)
				.temperature(0.7).build();

		String json = mapper.writeValueAsString(request);
		assertThat(json).contains("\"model\":\"gpt-4o-mini\"");
		assertThat(json).contains("\"max_tokens\":100");
		assertThat(json).contains("\"temperature\":0.7");
		assertThat(json).doesNotContain("\"stream\"");
		assertThat(json).doesNotContain("\"seed\"");
	}

	@Test
	void finishReason_unknownValue_deserializesToOther() throws Exception {
		String json = "\"new_unknown_reason\"";
		var reason = mapper.readValue(json, FinishReason.class);
		assertThat(reason).isEqualTo(FinishReason.OTHER);
	}

	@Test
	void usage_roundTrips() throws Exception {
		var usage = new Usage(100L, 50L, 150L);
		String json = mapper.writeValueAsString(usage);

		assertThat(json).contains("\"prompt_tokens\":100");
		assertThat(json).contains("\"completion_tokens\":50");
		assertThat(json).contains("\"total_tokens\":150");

		var decoded = mapper.readValue(json, Usage.class);
		assertThat(decoded).isEqualTo(usage);
	}

	@Test
	void embeddingRequest_singleInput_roundTrips() throws Exception {
		var request = new EmbeddingRequest("text-embedding-3-small", EmbeddingInput.of("Hello, world!"));
		String json = mapper.writeValueAsString(request);

		assertThat(json).contains("\"model\":\"text-embedding-3-small\"");
		assertThat(json).contains("\"Hello, world!\"");
	}

	@Test
	void embeddingRequest_multipleInput_roundTrips() throws Exception {
		var request = new EmbeddingRequest("text-embedding-3-small", EmbeddingInput.of(List.of("first", "second")));
		String json = mapper.writeValueAsString(request);

		assertThat(json).contains("\"first\"");
		assertThat(json).contains("\"second\"");
	}

	@Test
	void modelObject_roundTrips() throws Exception {
		String json = """
				{"id":"gpt-4o","object":"model","created":1712361441,"owned_by":"openai"}""";
		var model = mapper.readValue(json, ModelObject.class);

		SoftAssertions softly = new SoftAssertions();
		softly.assertThat(model.id()).isEqualTo("gpt-4o");
		softly.assertThat(model.object()).isEqualTo("model");
		softly.assertThat(model.ownedBy()).isEqualTo("openai");
		softly.assertAll();
	}

	@Test
	void toolChoice_auto_serializesAsString() throws Exception {
		String json = mapper.writeValueAsString(ToolChoice.AUTO);
		assertThat(json).isEqualTo("\"auto\"");
	}

	@Test
	void toolChoice_specific_serializesCorrectly() throws Exception {
		var choice = ToolChoice.forFunction("get_weather");
		String json = mapper.writeValueAsString(choice);
		assertThat(json).contains("\"type\":\"function\"");
		assertThat(json).contains("\"get_weather\"");
	}

	@Test
	void responseFormat_text_roundTrips() throws Exception {
		String json = mapper.writeValueAsString(ResponseFormat.TEXT);
		assertThat(json).contains("\"type\":\"text\"");
	}

	@Test
	void stopSequence_single_roundTrips() throws Exception {
		var stop = StopSequence.of("\\n\\n");
		String json = mapper.writeValueAsString(stop);
		assertThat(json).isEqualTo("\"\\n\\n\"");
	}

	@Test
	void stopSequence_multiple_roundTrips() throws Exception {
		var stop = StopSequence.of(List.of("stop1", "stop2"));
		String json = mapper.writeValueAsString(stop);
		assertThat(json).contains("\"stop1\"");
		assertThat(json).contains("\"stop2\"");
	}
}
