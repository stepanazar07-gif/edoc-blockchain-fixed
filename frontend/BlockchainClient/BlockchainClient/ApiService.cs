using System;
using System.Collections.Generic;
using System.IO;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using Microsoft.Maui.Storage;

namespace BlockchainClient
{
    public class ApiService
    {
        private readonly HttpClient _httpClient = new();
        private readonly string _baseUrl;

        private static readonly JsonSerializerOptions JsonOptions = new()
        {
            PropertyNameCaseInsensitive = true
        };

        public string LastError { get; private set; } = "";

        public ApiService(string? baseUrl = null)
        {
            if (!string.IsNullOrWhiteSpace(baseUrl))
            {
                _baseUrl = NormalizeBaseUrl(baseUrl);
                return;
            }

            var envBaseUrl = Environment.GetEnvironmentVariable("EDOC_API_URL");
            if (!string.IsNullOrWhiteSpace(envBaseUrl))
            {
                _baseUrl = NormalizeBaseUrl(envBaseUrl);
                return;
            }

            _baseUrl = "http://186.246.31.220:18080";
        }

        public async Task<string?> RegisterAsync(string name, int age, string phone, string password)
        {
            var data = new { name, age, phone, password };
            var response = await PostJsonAsync("/register", data, includeToken: false);
            return await ReadTokenAsync(response);
        }

        public async Task<string?> LoginAsync(string phone, string password)
        {
            var data = new { phone, password };
            var response = await PostJsonAsync("/login", data, includeToken: false);
            return await ReadTokenAsync(response);
        }

        public async Task<UserInfo?> GetCurrentUserAsync()
        {
            await AddTokenAsync();
            var response = await _httpClient.GetAsync($"{_baseUrl}/me");
            if (!response.IsSuccessStatusCode) return null;
            var json = await response.Content.ReadAsStringAsync();
            return JsonSerializer.Deserialize<UserInfo>(json, JsonOptions);
        }

        public async Task<byte[]?> GetAvatarAsync()
        {
            await AddTokenAsync();
            var response = await _httpClient.GetAsync($"{_baseUrl}/me/avatar");
            if (!response.IsSuccessStatusCode) return null;
            return await response.Content.ReadAsByteArrayAsync();
        }

        public async Task<bool> UploadAvatarAsync(string fileName, byte[] bytes)
        {
            await AddTokenAsync();

            using var form = new MultipartFormDataContent();
            var content = new ByteArrayContent(bytes);
            content.Headers.ContentType = new MediaTypeHeaderValue(GetImageMimeType(fileName));
            form.Add(content, "avatar", fileName);

            var response = await _httpClient.PostAsync($"{_baseUrl}/me/avatar", form);
            return response.IsSuccessStatusCode;
        }

        public async Task<List<UserInfo>?> GetAllUsersAsync(string? id = null)
        {
            await AddTokenAsync();
            var path = string.IsNullOrWhiteSpace(id)
                ? "/users"
                : $"/users?id={Uri.EscapeDataString(id.Trim())}";
            var response = await _httpClient.GetAsync($"{_baseUrl}{path}");
            if (!response.IsSuccessStatusCode) return null;
            var json = await response.Content.ReadAsStringAsync();
            return JsonSerializer.Deserialize<List<UserInfo>>(json, JsonOptions);
        }

        public async Task<List<MyDocument>?> GetMyDocumentsAsync()
        {
            await AddTokenAsync();
            var response = await _httpClient.GetAsync($"{_baseUrl}/my-documents");
            if (!response.IsSuccessStatusCode) return null;
            var json = await response.Content.ReadAsStringAsync();
            return JsonSerializer.Deserialize<List<MyDocument>>(json, JsonOptions);
        }

        public async Task<UploadResult?> UploadDocumentAsync(string fileName, byte[] fileBytes, string mimeType)
        {
            await AddTokenAsync();
            LastError = "";

            using var form = new MultipartFormDataContent();
            var fileContent = new ByteArrayContent(fileBytes);
            fileContent.Headers.ContentType = new MediaTypeHeaderValue(mimeType);
            form.Add(fileContent, "file", fileName);
            form.Add(new StringContent(fileName), "title");

            var response = await _httpClient.PostAsync($"{_baseUrl}/document", form);
            var body = await response.Content.ReadAsStringAsync();
            if (!response.IsSuccessStatusCode)
            {
                LastError = string.IsNullOrWhiteSpace(body) ? response.ReasonPhrase ?? "Upload failed" : body.Trim();
                return null;
            }
            return JsonSerializer.Deserialize<UploadResult>(body, JsonOptions);
        }

        public async Task<ShareResult?> ShareDocumentAsync(string fileId, string receiverId)
        {
            var data = new { file_id = fileId, receiver_id = receiverId };
            var response = await PostJsonAsync("/share-document", data);
            if (!response.IsSuccessStatusCode) return null;
            var json = await response.Content.ReadAsStringAsync();
            return JsonSerializer.Deserialize<ShareResult>(json, JsonOptions);
        }

        public async Task<List<IncomingTransfer>?> GetIncomingTransfersAsync()
        {
            await AddTokenAsync();
            var response = await _httpClient.GetAsync($"{_baseUrl}/incoming-transfers");
            if (!response.IsSuccessStatusCode) return null;
            var json = await response.Content.ReadAsStringAsync();
            return JsonSerializer.Deserialize<List<IncomingTransfer>>(json, JsonOptions);
        }

        public async Task<AcceptTransferResult?> AcceptTransferAsync(string transferId, string fileHash)
        {
            var data = new { transfer_id = transferId, file_hash = fileHash };
            var response = await PostJsonAsync("/accept-transfer", data);
            if (!response.IsSuccessStatusCode) return null;
            var json = await response.Content.ReadAsStringAsync();
            return JsonSerializer.Deserialize<AcceptTransferResult>(json, JsonOptions);
        }

        public async Task<bool> DeclineTransferAsync(string transferId)
        {
            var data = new { transfer_id = transferId };
            var response = await PostJsonAsync("/decline-transfer", data);
            return response.IsSuccessStatusCode;
        }

        public async Task<List<ReceivedFile>?> GetReceivedFilesAsync()
        {
            await AddTokenAsync();
            var response = await _httpClient.GetAsync($"{_baseUrl}/received-files");
            if (!response.IsSuccessStatusCode) return null;
            var json = await response.Content.ReadAsStringAsync();
            return JsonSerializer.Deserialize<List<ReceivedFile>>(json, JsonOptions);
        }

        public async Task<byte[]?> DownloadFileAsync(string fileHash)
        {
            await AddTokenAsync();
            var response = await _httpClient.GetAsync($"{_baseUrl}/download/{Uri.EscapeDataString(fileHash)}");
            if (!response.IsSuccessStatusCode) return null;
            return await response.Content.ReadAsByteArrayAsync();
        }

        public static string ComputeSha256(byte[] bytes)
        {
            return Convert.ToHexString(SHA256.HashData(bytes)).ToLowerInvariant();
        }

        public static string GuessMimeType(string fileName)
        {
            var ext = Path.GetExtension(fileName).ToLowerInvariant();
            return ext switch
            {
                ".pdf" => "application/pdf",
                ".txt" => "text/plain",
                ".doc" => "application/msword",
                ".docx" => "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
                ".xls" => "application/vnd.ms-excel",
                ".xlsx" => "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
                ".png" => "image/png",
                ".jpg" or ".jpeg" => "image/jpeg",
                _ => "application/octet-stream"
            };
        }

        private async Task<HttpResponseMessage> PostJsonAsync(string path, object data, bool includeToken = true)
        {
            if (includeToken)
            {
                await AddTokenAsync();
            }
            var content = new StringContent(JsonSerializer.Serialize(data), Encoding.UTF8, "application/json");
            return await _httpClient.PostAsync($"{_baseUrl}{path}", content);
        }

        private async Task AddTokenAsync()
        {
            var token = await SecureStorage.GetAsync("auth_token");
            if (string.IsNullOrWhiteSpace(token))
            {
                _httpClient.DefaultRequestHeaders.Authorization = null;
                return;
            }
            _httpClient.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", token);
        }

        private async Task<string?> ReadTokenAsync(HttpResponseMessage response)
        {
            var json = await response.Content.ReadAsStringAsync();
            LastError = "";

            if (!response.IsSuccessStatusCode)
            {
                LastError = string.IsNullOrWhiteSpace(json) ? response.ReasonPhrase ?? "Request failed" : json.Trim();
                return null;
            }

            try
            {
                using var doc = JsonDocument.Parse(json);
                return doc.RootElement.TryGetProperty("token", out var token)
                    ? token.GetString()
                    : null;
            }
            catch (JsonException ex)
            {
                LastError = ex.Message;
                return null;
            }
        }

        private static string GetImageMimeType(string fileName)
        {
            var ext = Path.GetExtension(fileName).ToLowerInvariant();
            return ext == ".png" ? "image/png" : "image/jpeg";
        }

        private static string NormalizeBaseUrl(string url)
        {
            return url.Trim().TrimEnd('/');
        }
    }
}
