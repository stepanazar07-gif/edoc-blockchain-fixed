using System;

namespace BlockchainClient
{
    public class DocumentRecord
    {
        public string FileName { get; set; } = "";
        public string Status { get; set; } = "";
        public string Details { get; set; } = "";
        public string FileHash { get; set; } = "";
        public string CounterpartyId { get; set; } = "";
        public DateTime Timestamp { get; set; }
    }
}
