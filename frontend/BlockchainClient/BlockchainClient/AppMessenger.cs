using System;

namespace BlockchainClient
{
    // Легковесный потокобезопасный диспетчер событий для синхронизации страниц без MessagingCenter
    public static class AppMessenger
    {
        public static event Action? OnFileUploaded;

        public static void NotifyFileUploaded()
        {
            OnFileUploaded?.Invoke();
        }
    }
}