using System;
using System.Collections.ObjectModel;
using Microsoft.Maui.Controls;

namespace BlockchainClient
{
    public partial class HistoryPage : ContentPage
    {
        public ObservableCollection<DocumentRecord> Records { get; } = new();

        public HistoryPage()
        {
            InitializeComponent();
            BindingContext = this;
        }

        protected override void OnAppearing()
        {
            base.OnAppearing();
            LoadHistory();
        }

        private void LoadHistory()
        {
            Records.Clear();
            foreach (var record in HistoryStorage.Load())
            {
                Records.Add(record);
            }
        }

        private void OnClearClicked(object sender, EventArgs e)
        {
            HistoryStorage.Save(new System.Collections.Generic.List<DocumentRecord>());
            LoadHistory();
        }
    }
}
