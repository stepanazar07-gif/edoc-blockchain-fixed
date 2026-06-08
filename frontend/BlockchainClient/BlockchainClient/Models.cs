using System;
using System.Text.Json.Serialization;

namespace BlockchainClient
{
    public class UserInfo
    {
        [JsonPropertyName("id")]
        public string Id { get; set; } = "";

        [JsonPropertyName("name")]
        public string Name { get; set; } = "";

        [JsonPropertyName("age")]
        public int Age { get; set; }

        [JsonPropertyName("phone")]
        public string Phone { get; set; } = "";

        [JsonPropertyName("avatar_url")]
        public string AvatarUrl { get; set; } = "";

        [JsonPropertyName("created_at")]
        public DateTime CreatedAt { get; set; }
    }

    public class MyDocument
    {
        [JsonPropertyName("id")]
        public string Id { get; set; } = "";

        [JsonPropertyName("owner_id")]
        public string OwnerId { get; set; } = "";

        [JsonPropertyName("file_name")]
        public string FileName { get; set; } = "";

        [JsonPropertyName("file_hash")]
        public string FileHash { get; set; } = "";

        [JsonPropertyName("file_size")]
        public long FileSize { get; set; }

        [JsonPropertyName("mime_type")]
        public string MimeType { get; set; } = "";

        [JsonPropertyName("upload_date")]
        public string UploadDate { get; set; } = "";

        [JsonPropertyName("upload_time")]
        public string UploadTime { get; set; } = "";

        [JsonPropertyName("uploaded_by")]
        public string UploadedBy { get; set; } = "";

        public string FileSizeText => FileSizeFormatter.FormatBytes(FileSize);
        public string ShortHash => FileHash.Length > 16 ? $"{FileHash[..16]}..." : FileHash;
    }

    public class IncomingTransfer
    {
        [JsonPropertyName("transfer_id")]
        public string TransferId { get; set; } = "";

        [JsonPropertyName("file_id")]
        public string FileId { get; set; } = "";

        [JsonPropertyName("file_name")]
        public string FileName { get; set; } = "";

        [JsonPropertyName("file_size")]
        public long FileSize { get; set; }

        [JsonPropertyName("mime_type")]
        public string MimeType { get; set; } = "";

        [JsonPropertyName("sender_id")]
        public string SenderId { get; set; } = "";

        [JsonPropertyName("receiver_id")]
        public string ReceiverId { get; set; } = "";

        [JsonPropertyName("status")]
        public string Status { get; set; } = "";

        [JsonPropertyName("transfer_date")]
        public string TransferDate { get; set; } = "";

        public string FileSizeText => FileSizeFormatter.FormatBytes(FileSize);
    }

    public class ReceivedFile : IncomingTransfer
    {
        [JsonPropertyName("file_hash")]
        public string FileHash { get; set; } = "";

        [JsonPropertyName("accepted_at")]
        public string AcceptedAt { get; set; } = "";

        public string ShortHash => FileHash.Length > 16 ? $"{FileHash[..16]}..." : FileHash;
    }

    public class UploadResult
    {
        [JsonPropertyName("transaction_hash")]
        public string TransactionHash { get; set; } = "";

        [JsonPropertyName("block_hash")]
        public string BlockHash { get; set; } = "";

        [JsonPropertyName("block_height")]
        public int BlockHeight { get; set; }

        [JsonPropertyName("document_id")]
        public string DocumentId { get; set; } = "";

        [JsonPropertyName("file_id")]
        public string FileId { get; set; } = "";

        [JsonPropertyName("file_name")]
        public string FileName { get; set; } = "";

        [JsonPropertyName("file_hash")]
        public string FileHash { get; set; } = "";

        [JsonPropertyName("file_size")]
        public long FileSize { get; set; }

        [JsonPropertyName("mime_type")]
        public string MimeType { get; set; } = "";

        [JsonPropertyName("uploaded_by")]
        public string UploadedBy { get; set; } = "";

        public string ShortHash => FileHash.Length > 16 ? $"{FileHash[..16]}..." : FileHash;
        public string FileSizeText => FileSizeFormatter.FormatBytes(FileSize);
    }

    public class ShareResult
    {
        [JsonPropertyName("transfer_id")]
        public string TransferId { get; set; } = "";

        [JsonPropertyName("file_hash")]
        public string FileHash { get; set; } = "";
    }

    public class AcceptTransferResult
    {
        [JsonPropertyName("transfer_id")]
        public string TransferId { get; set; } = "";

        [JsonPropertyName("file_id")]
        public string FileId { get; set; } = "";

        [JsonPropertyName("file_name")]
        public string FileName { get; set; } = "";

        [JsonPropertyName("file_hash")]
        public string FileHash { get; set; } = "";

        [JsonPropertyName("file_size")]
        public long FileSize { get; set; }

        [JsonPropertyName("mime_type")]
        public string MimeType { get; set; } = "";

        [JsonPropertyName("sender_id")]
        public string SenderId { get; set; } = "";

        [JsonPropertyName("status")]
        public string Status { get; set; } = "";
    }

    internal static class FileSizeFormatter
    {
        public static string FormatBytes(long bytes)
        {
            if (bytes < 1024) return $"{bytes} B";
            if (bytes < 1024 * 1024) return $"{bytes / 1024d:F1} KB";
            return $"{bytes / 1024d / 1024d:F1} MB";
        }
    }
}
