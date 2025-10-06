<!--  Sequence Diagram -->
https://www.mermaidchart.com/

sequenceDiagram
    actor Client
    actor Service as Service Provider
    participant CA as Client App
    participant SA as Service App
    participant GW as API Gateway
    participant US as User Service
    participant OS as Order Service
    participant NS as Notification Service
    participant LS as Location Service
    participant TS as Timer Service
    participant PS as Payment Service
    participant DB as Database

    Note over Client,DB: FASE 0: AUTHENTICATION & PROFILE
    Client->>CA: Buka app
    CA->>GW: Login request
    GW->>US: Authenticate user
    US->>DB: Verify credentials
    DB-->>US: User data
    US-->>GW: JWT Token + Profile
    GW-->>CA: Auth success
    CA->>Client: Show home screen
    
    Service->>SA: Buka app
    SA->>GW: Login request (provider)
    GW->>US: Authenticate provider
    US->>DB: Verify credentials
    DB-->>US: Provider data
    US-->>GW: JWT Token + Profile
    GW-->>SA: Auth success
    SA->>Service: Show dashboard

    Note over Client,DB: FASE 1: CREATE ORDER
    Client->>CA: Request create order
    CA->>GW: Create order (+ Auth Token)
    GW->>US: Validate token
    US-->>GW: Token valid
    GW->>OS: Create new order
    OS->>DB: Save order (status: PENDING)
    DB-->>OS: Order created
    
    Note over Client,DB: FASE 2: BROADCAST & MATCHING
    OS->>NS: Broadcast to nearby providers
    NS->>SA: Push notification (new order)
    SA->>Service: Show order notification
    Service->>SA: Accept order
    SA->>GW: Accept order request (+ Auth Token)
    GW->>US: Validate token
    US-->>GW: Token valid
    GW->>OS: Update order status (ACCEPTED)
    OS->>DB: Update order & assign provider
    OS->>NS: Notify client (order accepted)
    NS->>CA: Provider info & profile
    CA->>Client: Show matched provider
    
    Note over Client,DB: FASE 3: TRACKING
    Service->>SA: Start heading to location
    SA->>GW: Update status (ON_THE_WAY)
    GW->>OS: Update order status
    
    loop Every 5 seconds
        SA->>GW: Send GPS coordinates
        GW->>LS: Update location
        LS->>DB: Save location data
        LS->>CA: Real-time location
        CA->>Client: Show on map
    end
    
    Note over Client,DB: FASE 4: ARRIVAL
    Service->>SA: Click "Telah Sampai"
    SA->>GW: Update status (ARRIVED)
    GW->>OS: Update order status
    OS->>NS: Notify client (arrived)
    NS->>CA: Show notification
    CA->>Client: Service has arrived
    
    Note over Client,DB: FASE 5: WORK SESSION
    Service->>SA: Click "Mulai Bekerja"
    SA->>GW: Start work session
    GW->>TS: Start timer
    TS->>DB: Save session start time
    TS-->>SA: Timer started
    
    loop While working
        TS->>SA: Update duration display
        SA->>Service: Show elapsed time
    end
    
    Service->>SA: Click "Selesai"
    SA->>GW: End work session
    GW->>TS: Stop timer
    TS->>DB: Save session end time
    TS->>TS: Calculate duration
    
    Note over Client,DB: FASE 6: PAYMENT
    TS->>PS: Send duration data
    PS->>PS: Calculate cost<br/>(duration/5 * 1000)
    PS->>DB: Create payment record
    PS->>CA: Send bill details
    CA->>Client: Show payment amount
    
    Client->>CA: Confirm payment
    CA->>GW: Process payment
    GW->>PS: Execute payment
    PS->>PS: Process via payment gateway
    PS->>DB: Update payment status
    PS->>DB: Update wallet balances
    PS->>NS: Payment success notification
    NS->>SA: Payment received
    NS->>CA: Payment confirmed
    
    Note over Client,DB: FASE 7: RATING & COMPLETE
    CA->>Client: Request rating
    Client->>CA: Submit rating & review
    CA->>GW: Send rating
    GW->>OS: Update order (COMPLETED)
    OS->>DB: Save rating & complete order
    OS->>NS: Notify provider (new rating)
    NS->>SA: Show rating received
    
    CA->>Client: Show order summary
    SA->>Service: Show earnings

    Note over Client,DB: UPDATE PROFILE (Optional Flow)
    Client->>CA: Update profile request
    CA->>GW: Update profile data (+ Auth Token)
    GW->>US: Validate & update profile
    US->>DB: Update user data
    DB-->>US: Profile updated
    US-->>GW: Update success
    GW-->>CA: Profile updated
    CA->>Client: Show success message
    
    Service->>SA: Update profile request
    SA->>GW: Update profile data (+ Auth Token)
    GW->>US: Validate & update profile
    US->>DB: Update provider data
    DB-->>US: Profile updated
    US-->>GW: Update success
    GW-->>SA: Profile updated
    SA->>Service: Show success message

<!-- flowchart  -->
https://www.mermaidchart.com/

flowchart TD
    Start([Client Membuka App]) --> CheckAuth{User<br/>Terautentikasi?}
    
    subgraph AUTH["🔐 AUTHENTICATION FLOW"]
        direction TB
        CheckAuth -->|Tidak| LoginChoice{Pilih Metode<br/>Login}
        LoginChoice -->|Credential| LoginForm[Input Email & Password]
        LoginChoice -->|Google OAuth| GoogleAuth[Login dengan Google]
        LoginChoice -->|Belum Punya Akun| RegisterChoice{Pilih Metode<br/>Registrasi}
        
        RegisterChoice -->|Email/Phone| RegisterForm[Input:<br/>- Full Name<br/>- Email<br/>- Phone<br/>- Password<br/>- User Type]
        RegisterChoice -->|Google OAuth| GoogleAuth
        
        RegisterForm --> SaveUser[(Save to users table)]
        SaveUser --> SendOTP[Kirim OTP]
        SendOTP --> InputOTP[Input OTP Code]
        InputOTP --> VerifyOTP{OTP Valid?}
        VerifyOTP -->|Tidak| OTPRetry{Coba Lagi?}
        OTPRetry -->|Ya| InputOTP
        OTPRetry -->|Tidak| Start
        VerifyOTP -->|Ya| UpdateVerified[(is_verified = true)]
        UpdateVerified --> CreateProfile[(Create user_profiles)]
        CreateProfile --> GenerateToken[Generate JWT Token]
        
        LoginForm --> ValidateCredentials[(Check users table)]
        ValidateCredentials --> CheckLocked{Account Locked?}
        CheckLocked -->|Ya| LockedMsg[Account Terkunci]
        LockedMsg --> Start
        CheckLocked -->|Tidak| VerifyPassword{Password Benar?}
        VerifyPassword -->|Tidak| IncrementAttempt[(login_attempts + 1)]
        IncrementAttempt --> CheckMaxAttempt{Attempts >= 5?}
        CheckMaxAttempt -->|Ya| LockAccount[(Lock 30 min)]
        LockAccount --> LockedMsg
        CheckMaxAttempt -->|Tidak| LoginForm
        VerifyPassword -->|Ya| ResetAttempt[(Reset attempts)]
        
        GoogleAuth --> ValidateGoogle[(Verify Google Token)]
        ValidateGoogle --> CheckUserExist{User Exists?}
        CheckUserExist -->|Tidak| CreateGoogleUser[(Create user)]
        CheckUserExist -->|Ya| ResetAttempt
        CreateGoogleUser --> GenerateToken
        ResetAttempt --> GenerateToken
        GenerateToken --> SaveRefreshToken[(Save tokens)]
    end
    
    SaveRefreshToken --> CheckAuth
    CheckAuth -->|Ya| ValidateToken[(Validate JWT)]
    ValidateToken --> TokenValid{Token Valid?}
    TokenValid -->|Tidak| LoginChoice
    TokenValid -->|Ya| CheckUserType{User Type?}
    
    CheckUserType -->|CLIENT| ClientFlow
    CheckUserType -->|SERVICE_PROVIDER| ProviderFlow
    CheckUserType -->|ADMIN| AdminFlow
    
    subgraph ADMIN["👨‍💼 ADMIN DASHBOARD - Orange"]
        direction TB
        AdminFlow[Dashboard Admin]
        AdminFlow --> AdminMenu{Admin Menu}
        AdminMenu -->|Review Verifications| ReviewProviders[List Pending<br/>Verifications]
        ReviewProviders --> SelectVerification[Select Provider]
        SelectVerification --> ReviewDetail[Review Documents:<br/>- KTP<br/>- Certificates<br/>- Portfolio]
        ReviewDetail --> AdminDecision{Approve?}
        AdminDecision -->|Reject| InputRejectionReason[Input Reason]
        InputRejectionReason --> SaveRejection[(status: REJECTED)]
        SaveRejection --> NotifyRejection[(Notify Rejected)]
        NotifyRejection --> ReviewProviders
        AdminDecision -->|Approve| SaveApproval[(status: APPROVED<br/>is_verified: true)]
        SaveApproval --> NotifyApproval[(Notify Approved)]
        NotifyApproval --> ReviewProviders
        AdminMenu -->|Dashboard Stats| AdminFlow
        AdminMenu -->|Manage Users| AdminFlow
        AdminMenu -->|Manage Categories| AdminFlow
    end
    
    subgraph PROVIDER["🔧 SERVICE PROVIDER FLOW - Blue"]
        direction TB
        ProviderFlow --> CheckProviderVerified{Provider<br/>Verified?}
        
        CheckProviderVerified -->|Tidak| ProviderVerification[Verification Form:<br/>- KTP<br/>- Selfie + KTP<br/>- Certificates<br/>- Portfolio<br/>- Bank Info]
        ProviderVerification --> SubmitVerification[Submit]
        SubmitVerification --> SaveVerificationData[(Save verification)]
        SaveVerificationData --> WaitingApproval[Waiting Approval<br/>1-3 days]
        WaitingApproval --> CheckApprovalStatus{Status?}
        CheckApprovalStatus -->|Pending| WaitingApproval
        CheckApprovalStatus -->|Rejected| ShowRejectionReason[Show Rejection<br/>Reason]
        ShowRejectionReason --> ProviderVerification
        CheckApprovalStatus -->|Approved| UpdateProviderVerified[(is_verified = true)]
        UpdateProviderVerified --> NotifyApproved[(Notify Approved)]
        NotifyApproved --> DashboardProvider
        
        CheckProviderVerified -->|Ya| DashboardProvider[Dashboard Provider]
        DashboardProvider --> ProviderMenu{Provider Menu}
        ProviderMenu -->|Update Profile| UpdateProviderProfile[Edit Profile:<br/>- Bio<br/>- Service Types<br/>- Hourly Rate<br/>- Portfolio<br/>- Working Hours]
        UpdateProviderProfile --> SaveProviderProfile[(Update profile)]
        SaveProviderProfile --> DashboardProvider
        ProviderMenu -->|Set Availability| SetAvailability[Online/Offline<br/>Working Hours]
        SetAvailability --> SaveAvailability[(Save availability)]
        SaveAvailability --> DashboardProvider
        ProviderMenu -->|View Orders| ViewProviderOrders[Orders:<br/>- Pending Nearby<br/>- Active<br/>- History]
        ViewProviderOrders --> DashboardProvider
        ProviderMenu -->|Wallet| ViewProviderWallet[Wallet:<br/>- Balance<br/>- Withdrawal<br/>- Transactions]
        ViewProviderWallet --> DashboardProvider
        ProviderMenu -->|Statistics| ViewProviderStats[Stats:<br/>- Orders<br/>- Rating<br/>- Earnings]
        ViewProviderStats --> DashboardProvider
    end
    
    subgraph CLIENT["👤 CLIENT DASHBOARD - Green"]
        direction TB
        ClientFlow[Dashboard Client]
        ClientFlow --> ProfileMenu{Menu}
        ProfileMenu -->|Update Profile| UpdateClientProfile[Edit:<br/>- Name<br/>- Phone<br/>- Photo<br/>- Address]
        UpdateClientProfile --> SaveClientProfile[(Update profile)]
        SaveClientProfile --> ClientFlow
        ProfileMenu -->|My Orders| ViewClientOrders[Order History]
        ViewClientOrders --> ClientFlow
        ProfileMenu -->|Wallet| ViewClientWallet[Wallet Balance]
        ViewClientWallet --> ClientFlow
        ProfileMenu -->|Browse Providers| BrowseProviders[Provider List<br/>✓ Verified Only]
        BrowseProviders --> SelectProvider{Select?}
        SelectProvider -->|Tidak| ClientFlow
        SelectProvider -->|Ya| ViewProviderDetail[Provider Detail<br/>✓ Badge]
    end
    
    ViewProviderDetail --> ContactOption{Komunikasi?}
    ContactOption -->|Skip| CreateOrderFlow
    ContactOption -->|Chat| CommChat
    ContactOption -->|Voice| CommVoice
    ContactOption -->|Video| CommVideo
    
    subgraph COMM["💬 COMMUNICATION - Yellow OPTIONAL"]
        direction TB
        CommChat[Chat WebSocket]
        CommChat --> WSConnect[WebSocket Connect]
        WSConnect --> CreateChatRoom[(Create chat_rooms)]
        CreateChatRoom --> StartChat[Chat Active]
        StartChat --> SendMessage[Send Message]
        SendMessage --> WSPublish[WebSocket Publish]
        WSPublish --> RedisPublish[(Redis Pub/Sub)]
        RedisPublish --> SaveChatMsg[(Save message)]
        SaveChatMsg --> WSPushProvider[Push to Provider]
        WSPushProvider --> ChatContinue{Continue?}
        ChatContinue -->|Ya| SendMessage
        ChatContinue -->|Tidak| WSDisconnect[Disconnect]
        
        CommVoice[Voice Call WebRTC]
        CommVoice --> WSSignaling[WebSocket Signaling]
        WSSignaling --> CreateCallRoom[(Create call_logs)]
        CreateCallRoom --> SendOffer[WebRTC Offer]
        SendOffer --> ProviderResponse{Response?}
        ProviderResponse -->|Reject/Timeout| UpdateCallFailed[(Failed/Missed)]
        ProviderResponse -->|Accept| SendAnswer[WebRTC Answer]
        SendAnswer --> ICEExchange[ICE Exchange]
        ICEExchange --> STUNCheck{P2P?}
        STUNCheck -->|Ya| P2PConnection[STUN P2P]
        STUNCheck -->|Tidak| TURNRelay[TURN Relay]
        P2PConnection --> VoiceActive[Call Active]
        TURNRelay --> VoiceActive
        VoiceActive --> SaveCallComplete[(Save duration)]
        
        CommVideo[Video Call WebRTC]
        CommVideo --> WSVideoSignaling[WebSocket Signaling]
        WSVideoSignaling --> CreateVideoRoom[(Create call_logs)]
        CreateVideoRoom --> SendVideoOffer[WebRTC Offer]
        SendVideoOffer --> VideoResponse{Response?}
        VideoResponse -->|Reject/Timeout| UpdateVideoFailed[(Failed/Missed)]
        VideoResponse -->|Accept| SendVideoAnswer[WebRTC Answer]
        SendVideoAnswer --> VideoICEExchange[ICE Exchange]
        VideoICEExchange --> VideoSTUNCheck{P2P?}
        VideoSTUNCheck -->|Ya| VideoP2P[STUN P2P]
        VideoSTUNCheck -->|Tidak| VideoTURN[TURN Relay]
        VideoP2P --> VideoActive[Video Active]
        VideoTURN --> VideoActive
        VideoActive --> SaveVideoComplete[(Save duration)]
    end
    
    WSDisconnect --> AfterComm
    SaveCallComplete --> AfterComm
    UpdateCallFailed --> AfterComm
    SaveVideoComplete --> AfterComm
    UpdateVideoFailed --> AfterComm
    
    AfterComm{Deal?}
    AfterComm -->|Tidak| ViewProviderDetail
    AfterComm -->|Ya| CreateOrderFlow
    
    subgraph ORDER["📦 ORDER FLOW"]
        direction TB
        CreateOrderFlow[Create Order Form]
        CreateOrderFlow --> InputOrder[Input:<br/>- Category<br/>- Location<br/>- Description<br/>- Time]
        InputOrder --> ValidateTokenOrder[(Validate Token)]
        ValidateTokenOrder --> SubmitOrder[Submit]
        SubmitOrder --> SaveOrder[(Save order<br/>PENDING)]
        SaveOrder --> SaveTracking[(Create tracking)]
        SaveTracking --> Broadcast[Broadcast]
        Broadcast --> QueryProviders[(Get nearby providers)]
        QueryProviders --> SendPush[(Send notification)]
        SendPush --> WaitAccept{Accept?}
        WaitAccept -->|Timeout| CancelAuto[(CANCELLED)]
        CancelAuto --> NotifyCancelled[(Notify)]
        NotifyCancelled --> EndOrder1
        WaitAccept -->|Ya| ValidateProvider[(Validate Token)]
        ValidateProvider --> UpdateAccepted[(ACCEPTED)]
        UpdateAccepted --> NotifyClient[(Notify client)]
    end
    
    subgraph TRACKING["📍 TRACKING & WORK FLOW"]
        direction TB
        NotifyClient --> ProviderMove[ON_THE_WAY]
        ProviderMove --> UpdateOnWay[(Update status)]
        UpdateOnWay --> StartTracking[GPS Tracking]
        StartTracking --> TrackLoop{Active?}
        TrackLoop -->|Every 5s| SendGPS[(Update location)]
        SendGPS --> PushLocation[(Push to client)]
        PushLocation --> TrackLoop
        TrackLoop -->|Arrived| UpdateArrived[(ARRIVED)]
        UpdateArrived --> NotifyArrived[(Notify)]
        NotifyArrived --> StartWork[MULAI BEKERJA]
        StartWork --> CreateSession[(Create session)]
        CreateSession --> UpdateInProgress[(IN_PROGRESS)]
        UpdateInProgress --> Working[Working]
        Working --> TimerUpdate[Timer]
        TimerUpdate --> CheckDone{Done?}
        CheckDone -->|Belum| Working
        CheckDone -->|Ya| StopWork[SELESAI]
        StopWork --> EndSession[(End session)]
    end
    
    subgraph PAYMENT["💰 PAYMENT FLOW"]
        direction TB
        EndSession --> CalculateCost[Calculate Cost]
        CalculateCost --> UpdateOrderAmount[(Update amount)]
        UpdateOrderAmount --> CreatePayment[(Create payment)]
        CreatePayment --> ShowBill[Show Bill]
        ShowBill --> ClientPay[Confirm Payment]
        ClientPay --> ProcessPayment[(Process)]
        ProcessPayment --> PaymentCheck{Success?}
        PaymentCheck -->|Gagal| SavePaymentFailed[(FAILED)]
        SavePaymentFailed --> PaymentRetry[Retry]
        PaymentRetry --> ClientPay
        PaymentCheck -->|Berhasil| UpdatePaymentSuccess[(COMPLETED)]
        UpdatePaymentSuccess --> UpdateWallets[(Update wallets)]
        UpdateWallets --> NotifyPayment[(Notify)]
    end
    
    subgraph RATING["⭐ RATING FLOW"]
        direction TB
        NotifyPayment --> RequestRating[Request Rating]
        RequestRating --> ClientRate[Submit Rating]
        ClientRate --> SaveRating[(Save rating)]
        SaveRating --> UpdateCompleted[(COMPLETED)]
        UpdateCompleted --> UpdateProviderStats[(Update stats)]
        UpdateProviderStats --> NotifyRating[(Notify)]
        NotifyRating --> ShowSummary[Summary]
    end
    
    EndOrder1([End - Cancelled])
    ShowSummary --> EndOrder2([End - Success])
    
    style AUTH fill:#E8F5E9
    style ADMIN fill:#FFE0B2
    style PROVIDER fill:#BBDEFB
    style CLIENT fill:#C8E6C9
    style COMM fill:#FFF9C4
    style ORDER fill:#B2EBF2
    style TRACKING fill:#F8BBD0
    style PAYMENT fill:#D1C4E9
    style RATING fill:#FFCCBC
    
    style Start fill:#90EE90
    style EndOrder1 fill:#FFB6C1
    style EndOrder2 fill:#90EE90
    style GenerateToken fill:#FFD700
    style DashboardProvider fill:#87CEFA
    style ClientFlow fill:#98FB98
    style AdminFlow fill:#FFA07A

<!-- Diagranm Database Relation Micoservices -->

// Database Schema untuk Cleaning Service App
// Copy paste code ini ke https://dbdiagram.io/

// ==========================================
// DATABASE 1: USER SERVICE
// ==========================================
Table users {
  id uuid [primary key, default: `uuid_generate_v4()`]
  email varchar(255) [unique, not null]
  phone varchar(20) [unique, not null]
  password_hash varchar(255) [not null]
  full_name varchar(255) [not null]
  user_type varchar(20) [not null, note: 'CLIENT, SERVICE_PROVIDER, ADMIN']
  profile_photo text
  date_of_birth date
  gender varchar(10) [note: 'MALE, FEMALE']
  is_active boolean [default: true]
  is_verified boolean [default: false, note: 'Email verified for all, KYC verified for providers']
  otp_code varchar(6)
  otp_expires_at timestamp
  last_login timestamp
  login_attempts int [default: 0]
  locked_until timestamp
  google_id varchar(255) [unique, note: 'Google OAuth ID']
  login_type varchar(20) [default: 'CREDENTIAL', note: 'CREDENTIAL, GOOGLE']
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  
  indexes {
    email
    phone
    user_type
    google_id
    is_verified
  }
}

Table user_profiles {
  id uuid [primary key, default: `uuid_generate_v4()`]
  user_id uuid [ref: > users.id, not null, unique]
  bio text
  service_types json [note: 'Array of service types for providers']
  hourly_rate decimal(10,2)
  total_orders int [default: 0]
  total_completed_orders int [default: 0]
  average_rating decimal(3,2) [default: 0]
  total_ratings int [default: 0]
  certifications json [note: 'Array of certification objects']
  working_hours json [note: 'Schedule availability object']
  service_areas json [note: 'Array of service area objects with city/province']
  is_available boolean [default: false, note: 'Provider online/offline status']
  portfolio_urls json [note: 'Array of portfolio image URLs']
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
}

Table provider_verifications {
  id uuid [primary key, default: `uuid_generate_v4()`]
  user_id uuid [ref: > users.id, not null, unique]
  id_card_number varchar(50) [not null]
  id_card_photo_url text [not null, note: 'KTP/ID Card photo']
  selfie_with_id_url text [not null, note: 'Selfie holding ID card']
  certificate_urls json [note: 'Array of certificate photo URLs']
  portfolio_urls json [note: 'Array of portfolio/work photos']
  bank_account_name varchar(255) [not null]
  bank_name varchar(100) [not null]
  bank_account_number varchar(50) [not null]
  status varchar(20) [not null, default: 'PENDING', note: 'PENDING, APPROVED, REJECTED']
  rejection_reason text
  verified_by uuid [note: 'Admin user_id who verified']
  verified_at timestamp
  submitted_at timestamp [default: `now()`]
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  
  indexes {
    user_id
    status
    submitted_at
    verified_at
  }
}

Table user_locations {
  id uuid [primary key, default: `uuid_generate_v4()`]
  user_id uuid [ref: > users.id, not null]
  label varchar(50) [note: 'Home, Office, etc']
  latitude decimal(10,8) [not null]
  longitude decimal(11,8) [not null]
  address text [not null]
  city varchar(100)
  province varchar(100)
  postal_code varchar(10)
  is_primary boolean [default: false]
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  
  indexes {
    user_id
    is_primary
    (latitude, longitude)
  }
}

Table refresh_tokens {
  id uuid [primary key, default: `uuid_generate_v4()`]
  user_id uuid [ref: > users.id, not null]
  token text [not null, unique]
  expires_at timestamp [not null]
  is_revoked boolean [default: false]
  revoked_at timestamp
  user_agent text
  ip_address varchar(45)
  created_at timestamp [default: `now()`]
  
  indexes {
    user_id
    token
    expires_at
    is_revoked
  }
}

Table fcm_tokens {
  id uuid [primary key, default: `uuid_generate_v4()`]
  user_id uuid [ref: > users.id, not null]
  device_id varchar(255) [not null]
  fcm_token text [not null]
  device_type varchar(20) [note: 'ANDROID, IOS, WEB']
  device_name varchar(255)
  is_active boolean [default: true]
  last_used timestamp [default: `now()`]
  created_at timestamp [default: `now()`]
  
  indexes {
    user_id
    device_id
    is_active
  }
}


// ==========================================
// DATABASE 2: ORDER SERVICE
// ==========================================
Table service_categories {
  id uuid [primary key, default: `uuid_generate_v4()`]
  name varchar(100) [not null]
  slug varchar(100) [unique, not null]
  description text
  icon varchar(255)
  base_rate_per_5_minutes decimal(10,2) [not null, note: 'Default: 1000']
  is_active boolean [default: true]
  display_order int [default: 0]
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  
  indexes {
    slug
    is_active
    display_order
  }
}

Table orders {
  id uuid [primary key, default: `uuid_generate_v4()`]
  order_number varchar(50) [unique, not null]
  client_id uuid [not null, note: 'Reference to User Service - NO FK']
  service_provider_id uuid [note: 'Reference to User Service - NO FK']
  service_category_id uuid [ref: > service_categories.id, note: 'NULL if custom category']
  custom_category_name varchar(100) [note: 'For custom service category']
  status varchar(20) [not null, note: 'PENDING, ACCEPTED, ON_THE_WAY, ARRIVED, IN_PROGRESS, COMPLETED, CANCELLED']
  description text
  service_latitude decimal(10,8) [not null, note: 'Client service location latitude']
  service_longitude decimal(11,8) [not null, note: 'Client service location longitude']
  service_address text [not null, note: 'Client service location address']
  requested_time timestamp [not null]
  broadcast_time timestamp [note: 'When order was broadcasted to providers']
  accepted_time timestamp
  arrived_time timestamp
  started_time timestamp
  completed_time timestamp
  cancelled_time timestamp
  duration_minutes int [default: 0]
  base_amount decimal(10,2) [default: 0]
  service_fee decimal(10,2) [default: 0]
  total_amount decimal(10,2) [default: 0]
  cancellation_reason text
  cancelled_by uuid [note: 'Reference to User Service - NO FK']
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  
  indexes {
    order_number
    client_id
    service_provider_id
    status
    created_at
    requested_time
  }
}

Table order_broadcasts {
  id uuid [primary key, default: `uuid_generate_v4()`]
  order_id uuid [ref: > orders.id, not null]
  provider_id uuid [not null, note: 'Reference to User Service - NO FK']
  notified_at timestamp [default: `now()`]
  seen_at timestamp
  is_accepted boolean [default: false]
  
  indexes {
    order_id
    provider_id
    notified_at
  }
}

Table work_sessions {
  id uuid [primary key, default: `uuid_generate_v4()`]
  order_id uuid [ref: > orders.id, not null, unique]
  start_time timestamp [not null]
  end_time timestamp
  pause_times json [note: 'Array of pause/resume timestamps']
  duration_minutes int [default: 0]
  status varchar(20) [not null, note: 'STARTED, PAUSED, COMPLETED']
  provider_notes text
  client_notes text
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  
  indexes {
    order_id
    status
    start_time
  }
}

Table ratings {
  id uuid [primary key, default: `uuid_generate_v4()`]
  order_id uuid [ref: > orders.id, not null, unique]
  client_id uuid [not null, note: 'Reference to User Service - NO FK']
  service_provider_id uuid [not null, note: 'Reference to User Service - NO FK']
  rating int [not null, note: '1-5 stars']
  review text
  rating_details json [note: 'Object: {cleanliness: 5, punctuality: 4, professionalism: 5}']
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  
  indexes {
    order_id
    service_provider_id
    client_id
    rating
    created_at
  }
}


// ==========================================
// DATABASE 3: LOCATION SERVICE
// ==========================================
Table order_tracking {
  id uuid [primary key, default: `uuid_generate_v4()`]
  order_id uuid [not null, unique, note: 'Reference to Order Service - NO FK']
  service_provider_id uuid [not null, note: 'Reference to User Service - NO FK']
  current_latitude decimal(10,8)
  current_longitude decimal(11,8)
  distance_km decimal(6,2)
  estimated_arrival_minutes int
  tracking_status varchar(20) [default: 'ACTIVE', note: 'ACTIVE, PAUSED, STOPPED']
  last_updated timestamp [default: `now()`]
  created_at timestamp [default: `now()`]
  
  indexes {
    order_id
    service_provider_id
    tracking_status
    last_updated
  }
}

Table location_history {
  id uuid [primary key, default: `uuid_generate_v4()`]
  order_id uuid [not null, note: 'Reference to Order Service - NO FK']
  service_provider_id uuid [not null, note: 'Reference to User Service - NO FK']
  latitude decimal(10,8) [not null]
  longitude decimal(11,8) [not null]
  speed_kmh decimal(5,2)
  accuracy_meters int
  heading_degrees int [note: 'Compass direction 0-360']
  recorded_at timestamp [default: `now()`]
  
  indexes {
    order_id
    service_provider_id
    recorded_at
  }
}


// ==========================================
// DATABASE 4: PAYMENT SERVICE
// ==========================================
Table payments {
  id uuid [primary key, default: `uuid_generate_v4()`]
  order_id uuid [not null, unique, note: 'Reference to Order Service - NO FK']
  payment_number varchar(50) [unique, not null]
  client_id uuid [not null, note: 'Reference to User Service - NO FK']
  service_provider_id uuid [not null, note: 'Reference to User Service - NO FK']
  base_amount decimal(10,2) [not null]
  service_fee decimal(10,2) [default: 0]
  platform_fee decimal(10,2) [default: 0, note: 'Platform commission']
  discount decimal(10,2) [default: 0]
  total_amount decimal(10,2) [not null]
  provider_earnings decimal(10,2) [not null, note: 'Amount received by provider']
  payment_method varchar(20) [not null, note: 'CASH, WALLET, BANK_TRANSFER, E_WALLET']
  payment_gateway varchar(50) [note: 'midtrans, xendit, etc']
  external_payment_id varchar(255) [note: 'Payment gateway transaction ID']
  status varchar(20) [not null, note: 'PENDING, PROCESSING, COMPLETED, FAILED, REFUNDED']
  payment_proof text
  notes text
  paid_at timestamp
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  
  indexes {
    payment_number
    order_id
    client_id
    service_provider_id
    status
    external_payment_id
  }
}

Table payment_history {
  id uuid [primary key, default: `uuid_generate_v4()`]
  payment_id uuid [ref: > payments.id, not null]
  status varchar(20) [not null]
  message text
  metadata json
  created_at timestamp [default: `now()`]
  
  indexes {
    payment_id
    created_at
  }
}

Table wallets {
  id uuid [primary key, default: `uuid_generate_v4()`]
  user_id uuid [not null, unique, note: 'Reference to User Service - NO FK']
  balance decimal(15,2) [default: 0, not null]
  pending_balance decimal(15,2) [default: 0, note: 'Orders in progress']
  frozen_balance decimal(15,2) [default: 0, note: 'Suspended/under review']
  currency varchar(3) [default: 'IDR']
  is_active boolean [default: true]
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  
  indexes {
    user_id
    is_active
  }
}

Table wallet_transactions {
  id uuid [primary key, default: `uuid_generate_v4()`]
  wallet_id uuid [ref: > wallets.id, not null]
  transaction_number varchar(50) [unique, not null]
  type varchar(10) [not null, note: 'CREDIT, DEBIT']
  category varchar(20) [not null, note: 'TOP_UP, PAYMENT, REFUND, WITHDRAWAL, COMMISSION, FEE']
  amount decimal(15,2) [not null]
  balance_before decimal(15,2) [not null]
  balance_after decimal(15,2) [not null]
  reference_id uuid [note: 'order_id or payment_id']
  reference_type varchar(20) [note: 'ORDER, PAYMENT, TOP_UP, WITHDRAWAL']
  description text
  metadata json
  created_at timestamp [default: `now()`]
  
  indexes {
    wallet_id
    transaction_number
    type
    category
    reference_id
    created_at
  }
}

Table withdrawal_requests {
  id uuid [primary key, default: `uuid_generate_v4()`]
  user_id uuid [not null, note: 'Reference to User Service - NO FK']
  wallet_id uuid [ref: > wallets.id, not null]
  withdrawal_number varchar(50) [unique, not null]
  amount decimal(15,2) [not null]
  bank_name varchar(100) [not null]
  bank_account_number varchar(50) [not null]
  bank_account_name varchar(255) [not null]
  status varchar(20) [not null, default: 'PENDING', note: 'PENDING, PROCESSING, COMPLETED, REJECTED']
  admin_notes text
  processed_by uuid [note: 'Admin user_id']
  processed_at timestamp
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  
  indexes {
    user_id
    wallet_id
    withdrawal_number
    status
    created_at
  }
}


// ==========================================
// DATABASE 5: COMMUNICATION SERVICE (OPTIONAL)
// Technology Stack:
// - WebSocket (Socket.IO) for Chat & Signaling
// - Redis Pub/Sub for scaling across instances
// - WebRTC for P2P media streaming
// - Coturn (STUN/TURN) for NAT traversal
// ==========================================

Table websocket_sessions {
  id uuid [primary key, default: `uuid_generate_v4()`]
  user_id uuid [not null, note: 'Reference to User Service - NO FK']
  socket_id varchar(255) [not null, unique]
  connection_type varchar(20) [not null, note: 'CHAT, VOICE_SIGNALING, VIDEO_SIGNALING']
  is_connected boolean [default: true]
  user_agent text
  ip_address varchar(45)
  connected_at timestamp [default: `now()`]
  disconnected_at timestamp
  
  indexes {
    user_id
    socket_id
    is_connected
    connection_type
  }
}

Table chat_rooms {
  id uuid [primary key, default: `uuid_generate_v4()`]
  client_id uuid [not null, note: 'Reference to User Service - NO FK']
  service_provider_id uuid [not null, note: 'Reference to User Service - NO FK']
  order_id uuid [note: 'Reference to Order Service - NO FK if order created']
  websocket_room_id varchar(255) [not null, unique, note: 'Socket.IO room ID']
  last_message text
  last_message_at timestamp
  unread_count_client int [default: 0]
  unread_count_provider int [default: 0]
  is_active boolean [default: true]
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  
  indexes {
    client_id
    service_provider_id
    order_id
    is_active
    last_message_at
  }
}

Table chat_messages {
  id uuid [primary key, default: `uuid_generate_v4()`]
  chat_room_id uuid [ref: > chat_rooms.id, not null]
  sender_id uuid [not null, note: 'Reference to User Service - NO FK']
  receiver_id uuid [not null, note: 'Reference to User Service - NO FK']
  message_type varchar(20) [not null, note: 'TEXT, IMAGE, FILE, LOCATION, AUDIO']
  message_content text [not null]
  file_url text [note: 'For image/file/audio messages']
  file_size_kb int
  latitude decimal(10,8) [note: 'For location messages']
  longitude decimal(11,8) [note: 'For location messages']
  is_read boolean [default: false]
  read_at timestamp
  is_deleted boolean [default: false]
  created_at timestamp [default: `now()`]
  
  indexes {
    chat_room_id
    sender_id
    receiver_id
    is_read
    created_at
  }
}

Table call_logs {
  id uuid [primary key, default: `uuid_generate_v4()`]
  caller_id uuid [not null, note: 'Reference to User Service - NO FK']
  receiver_id uuid [not null, note: 'Reference to User Service - NO FK']
  order_id uuid [note: 'Reference to Order Service - NO FK if related']
  call_type varchar(10) [not null, note: 'VOICE, VIDEO']
  status varchar(20) [not null, note: 'RINGING, ONGOING, COMPLETED, MISSED, REJECTED, FAILED']
  duration_seconds int [default: 0]
  webrtc_session_id varchar(255)
  ice_servers_used json [note: 'STUN/TURN servers used']
  connection_quality varchar(20) [note: 'EXCELLENT, GOOD, FAIR, POOR']
  started_at timestamp [not null]
  answered_at timestamp
  ended_at timestamp
  created_at timestamp [default: `now()`]
  
  indexes {
    caller_id
    receiver_id
    order_id
    call_type
    status
    started_at
  }
}

Table communication_media {
  id uuid [primary key, default: `uuid_generate_v4()`]
  chat_room_id uuid [ref: > chat_rooms.id, note: 'NULL if from call']
  call_log_id uuid [ref: > call_logs.id, note: 'NULL if from chat']
  uploader_id uuid [not null, note: 'Reference to User Service - NO FK']
  media_type varchar(20) [not null, note: 'IMAGE, VIDEO, AUDIO, DOCUMENT']
  file_name varchar(255) [not null]
  file_url text [not null]
  thumbnail_url text
  file_size_kb int
  mime_type varchar(100)
  duration_seconds int [note: 'For video/audio']
  created_at timestamp [default: `now()`]
  
  indexes {
    chat_room_id
    call_log_id
    uploader_id
    media_type
    created_at
  }
}


// ==========================================
// DATABASE 6: NOTIFICATION SERVICE
// ==========================================
Table notifications {
  id uuid [primary key, default: `uuid_generate_v4()`]
  user_id uuid [not null, note: 'Reference to User Service - NO FK']
  title varchar(255) [not null]
  message text [not null]
  type varchar(20) [not null, note: 'ORDER, PAYMENT, PROMOTION, SYSTEM, VERIFICATION, RATING']
  data json [note: 'Additional data like order_id, payment_id, etc']
  action_url text [note: 'Deep link URL']
  is_read boolean [default: false]
  is_push_sent boolean [default: false]
  push_sent_at timestamp
  read_at timestamp
  created_at timestamp [default: `now()`]
  
  indexes {
    user_id
    type
    is_read
    created_at
  }
}

Table notification_templates {
  id uuid [primary key, default: `uuid_generate_v4()`]
  code varchar(50) [unique, not null, note: 'ORDER_CREATED, ORDER_ACCEPTED, PAYMENT_SUCCESS, etc']
  title_template varchar(255) [not null]
  message_template text [not null]
  type varchar(20) [not null]
  variables json [note: 'Array of variable names used in template']
  is_active boolean [default: true]
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  
  indexes {
    code
    type
    is_active
  }
}


Back-end Microservices
│   ├── 📂 api-gateway/             # API Gateway (GW)
│   │   ├── 📂 cmd/
│   │   ├── 📂 internal/
│   │   ├── 📜 go.mod
│   │   └── 📜 Dockerfile
│   │
│   └── 📂 services/                # Kumpulan semua microservices
│       ├── 📂 user-service/        # User Service
│       │   ├── 📂 cmd/
│       │   ├── 📂 internal/
│       │   │   ├── cache/
│       │   │   ├── consumers/
│       │   │   ├── events/
│       │   │   ├── handlers/
│       │   │   ├── models/
│       │   │   ├── repository/
│       │   │   └── services/
│       │   ├── 📜 .env
│       │   ├── 📜 Dockerfile
│       │   ├── 📜 go.mod
│       │   └── 📜 README.md
│       │
│       ├── 📂 order-service/       # Order Service (OS)
│       │   ├── 📂 cmd/
│       │   ├── 📂 internal/
│       │   │   ├── cache/
│       │   │   ├── consumers/
│       │   │   ├── events/
│       │   │   ├── handlers/
│       │   │   ├── models/
│       │   │   ├── repository/
│       │   │   └── services/
│       │   ├── 📜 .env
│       │   ├── 📜 Dockerfile
│       │   ├── 📜 go.mod
│       │   └── 📜 README.md
│       │
│       ├── 📂 notification-service/ # Notification Service (NS) (skip  )
│       │   ├── 📂 cmd/
│       │   ├── 📂 internal/
│       │   │   ├── cache/
│       │   │   ├── consumers/
│       │   │   ├── events/
│       │   │   ├── handlers/
│       │   │   ├── models/
│       │   │   ├── repository/
│       │   │   └── services/
│       │   ├── 📜 .env
│       │   ├── 📜 Dockerfile
│       │   └── 📜 go.mod
│       │
│       ├── 📂 location-service/    # Location Service (LS)
│       │   ├── 📂 cmd/
│       │   ├── 📂 internal/
│       │   │   ├── cache/
│       │   │   ├── consumers/
│       │   │   ├── events/
│       │   │   ├── handlers/
│       │   │   ├── models/
│       │   │   ├── repository/
│       │   │   └── services/
│       │   ├── 📜 .env
│       │   ├── 📜 Dockerfile
│       │   └── 📜 go.mod
│       │
│       ├── 📂 timer-service/       # Timer Service (TS)
│       │   ├── 📂 cmd/
│       │   ├── 📂 internal/
│       │   │   ├── cache/
│       │   │   ├── consumers/
│       │   │   ├── events/
│       │   │   ├── handlers/
│       │   │   ├── models/
│       │   │   ├── repository/
│       │   │   └── services/
│       │   ├── 📜 .env
│       │   ├── 📜 Dockerfile
│       │   └── 📜 go.mod
│       │
│       └── 📂 payment-service/     # Payment Service (PS)
│           ├── 📂 cmd/
│           ├── 📂 internal/
│           │   ├── cache/
│           │   ├── consumers/
│           │   ├── events/
│           │   ├── handlers/
│           │   ├── models/
│           │   ├── repository/
│           │   └── services/
│           ├── 📜 .env
│           ├── 📜 Dockerfile
│           └── 📜 go.mod
│
├── 📂 database/                    # Skrip untuk Database init ke compose (DB)
│   
│   
│
├── 📜 .gitignore
├── 📜 docker-compose.yml           # Untuk menjalankan semua service bersamaan
└── 📜 README.md

