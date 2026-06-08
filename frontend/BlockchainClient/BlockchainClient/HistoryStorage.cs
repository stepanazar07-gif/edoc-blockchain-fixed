using System;
using System.Collections.Generic;
using System.Text.Json;
using Microsoft.Maui.Storage;

namespace BlockchainClient
{
    public static class HistoryStorage
    {
        private const string Key = "file_transfer_history_v2";

        public static List<DocumentRecord> Load()
        {
            var json = Preferences.Get(Key, string.Empty);
            if (string.IsNullOrWhiteSpace(json))
                return new List<DocumentRecord>();
            try
            {
                return JsonSerializer.Deserialize<List<DocumentRecord>>(json) ?? new List<DocumentRecord>();
            }
            catch
            {
                return new List<DocumentRecord>();
            }
        }

        public static void Save(List<DocumentRecord> records)
        {
            var json = JsonSerializer.Serialize(records);
            Preferences.Set(Key, json);
        }

        public static void Add(DocumentRecord record)
        {
            var list = Load();
            list.Insert(0, record);
            if (list.Count > 50) list.RemoveAt(list.Count - 1);
            Save(list);
        }
    }
}
